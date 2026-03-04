package mbserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testNetListener struct {
	acceptFn    func() (net.Conn, error)
	closeCalled bool
}

type testReadResult struct {
	data  []byte
	err   error
	after func()
}

type testNetConn struct {
	mu              sync.Mutex
	readResults     []testReadResult
	readIndex       int
	readCalls       int
	closed          bool
	readDeadlineErr error
}

func (l *testNetListener) Accept() (net.Conn, error) {
	return l.acceptFn()
}

func (l *testNetListener) Close() error {
	l.closeCalled = true
	return nil
}

func (l *testNetListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
}

func (c *testNetConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readCalls++

	if c.readIndex >= len(c.readResults) {
		return 0, io.EOF
	}

	result := c.readResults[c.readIndex]
	c.readIndex++

	n := copy(b, result.data)
	if result.after != nil {
		result.after()
	}

	return n, result.err
}

func (c *testNetConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *testNetConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *testNetConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
}

func (c *testNetConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5678}
}

func (c *testNetConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *testNetConn) SetReadDeadline(_ time.Time) error {
	return c.readDeadlineErr
}

func (c *testNetConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

type testTimeoutError struct{}

func (e testTimeoutError) Error() string   { return "timeout" }
func (e testTimeoutError) Timeout() bool   { return true }
func (e testTimeoutError) Temporary() bool { return true }

func TestListenTCP(t *testing.T) {
	s := NewServer()
	t.Cleanup(s.Shutdown)

	err := s.ListenTCP("127.0.0.1:0")
	require.NoError(t, err)
	require.Len(t, s.listeners, 1)
}

func TestListenTLS(t *testing.T) {
	s := NewServer()
	t.Cleanup(s.Shutdown)

	err := s.ListenTLS("127.0.0.1:0", newTestTLSConfig(t))
	require.NoError(t, err)
	require.Len(t, s.listeners, 1)
}

func TestAccept(t *testing.T) {
	t.Run("returns wrapped accept error", func(t *testing.T) {
		s := NewServer()
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			return nil, errors.New("accept failed")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unable to accept connections")
		assert.True(t, listener.closeCalled)
	})

	t.Run("handles timeout then returns error", func(t *testing.T) {
		s := NewServer()
		attempt := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			attempt++
			if attempt == 1 {
				return nil, testTimeoutError{}
			}
			return nil, errors.New("accept failed after timeout")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.GreaterOrEqual(t, attempt, 2)
	})

	t.Run("returns nil when closed before loop", func(t *testing.T) {
		s := NewServer()
		close(s.closeSignalChan)

		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			t.Fatal("accept should not be called after close signal")
			return nil, nil
		}}

		require.NoError(t, s.accept(listener))
		assert.True(t, listener.closeCalled)
	})

	t.Run("reads fragmented packet and enqueues request", func(t *testing.T) {
		s := NewServer()

		packet := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0x00, 0x00, 0x01}
		conn := &testNetConn{
			readResults: []testReadResult{
				{data: packet[:3]},
				{data: packet[3:7]},
				{data: packet[7:9]},
				{data: packet[9:]},
				{err: io.EOF},
			},
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			time.Sleep(20 * time.Millisecond)
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)

		require.Eventually(t, func() bool {
			return len(s.requestChan) == 1
		}, 500*time.Millisecond, 10*time.Millisecond)

		request := <-s.requestChan
		require.NotNil(t, request)
		assert.EqualValues(t, 0x03, request.frame.GetFunction())
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x01}, request.frame.GetData())
		assert.True(t, conn.closed)
	})

	t.Run("retries after read timeout", func(t *testing.T) {
		s := NewServer()

		packet := []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x04, 0x00, 0x01, 0x00, 0x01}
		conn := &testNetConn{
			readResults: []testReadResult{
				{err: testTimeoutError{}},
				{data: packet[:7]},
				{data: packet[7:]},
				{err: io.EOF},
			},
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			time.Sleep(20 * time.Millisecond)
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)

		require.Eventually(t, func() bool {
			return len(s.requestChan) == 1
		}, 500*time.Millisecond, 10*time.Millisecond)
		assert.True(t, conn.closed)
	})

	t.Run("stops connection loop when timeout occurs after close signal", func(t *testing.T) {
		s := NewServer()

		conn := &testNetConn{
			readResults: []testReadResult{{
				err: testTimeoutError{},
				after: func() {
					close(s.closeSignalChan)
				},
			}},
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			<-s.closeSignalChan
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.NoError(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		assert.True(t, conn.closed)
	})

	t.Run("drops malformed TCP frame", func(t *testing.T) {
		s := NewServer()

		packet := []byte{0x00, 0x04, 0x00, 0x01, 0x00, 0x06, 0xFF, 0x03, 0x00, 0x00, 0x00, 0x01}
		conn := &testNetConn{
			readResults: []testReadResult{
				{data: packet[:7]},
				{data: packet[7:]},
				{err: io.EOF},
			},
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		require.Eventually(t, func() bool {
			return conn.closed
		}, 500*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("drops request when queue is full and close signal is set", func(t *testing.T) {
		s := NewServer()

		for i := 0; i < cap(s.requestChan); i++ {
			s.requestChan <- &Request{}
		}

		packet := []byte{0x00, 0x03, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0x02, 0x00, 0x01}
		conn := &testNetConn{
			readResults: []testReadResult{
				{data: packet[:7]},
				{data: packet[7:], after: func() {
					close(s.closeSignalChan)
				}},
			},
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			<-s.closeSignalChan
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.NoError(t, err)
		assert.Equal(t, cap(s.requestChan), len(s.requestChan))
		assert.True(t, conn.closed)
	})

	t.Run("returns on header read non-timeout error", func(t *testing.T) {
		s := NewServer()

		conn := &testNetConn{readResults: []testReadResult{{err: errors.New("header read failed")}}}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		require.Eventually(t, func() bool {
			return conn.closed
		}, 500*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("returns on data read non-timeout error", func(t *testing.T) {
		s := NewServer()

		headerOnly := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF}
		conn := &testNetConn{readResults: []testReadResult{
			{data: headerOnly},
			{err: errors.New("data read failed")},
		}}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		require.Eventually(t, func() bool {
			return conn.closed
		}, 500*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("returns when data read timeout happens after close signal", func(t *testing.T) {
		s := NewServer()

		headerOnly := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF}
		conn := &testNetConn{readResults: []testReadResult{
			{data: headerOnly},
			{err: testTimeoutError{}, after: func() {
				close(s.closeSignalChan)
			}},
		}}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			<-s.closeSignalChan
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.NoError(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		assert.True(t, conn.closed)
	})

	t.Run("returns when SetReadDeadline fails", func(t *testing.T) {
		s := NewServer()

		validPacket := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0xFF, 0x03, 0x00, 0x00, 0x00, 0x01}
		conn := &testNetConn{
			readResults:     []testReadResult{{data: validPacket[:7]}, {data: validPacket[7:]}},
			readDeadlineErr: errors.New("deadline failed"),
		}

		acceptCalls := 0
		listener := &testNetListener{acceptFn: func() (net.Conn, error) {
			acceptCalls++
			if acceptCalls == 1 {
				return conn, nil
			}
			return nil, errors.New("stop accept")
		}}

		err := s.accept(listener)
		require.Error(t, err)
		assert.Never(t, func() bool {
			return len(s.requestChan) > 0
		}, 100*time.Millisecond, 10*time.Millisecond)
		require.Eventually(t, func() bool {
			return conn.closed
		}, 500*time.Millisecond, 10*time.Millisecond)
	})
}

func newTestTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{certificateDER},
			PrivateKey:  privateKey,
		}},
	}
}
