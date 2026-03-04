package mbserver

import (
	"errors"
	"io"
	"testing"

	"github.com/goburrow/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serialReadStep struct {
	data  []byte
	err   error
	after func()
}

type testSerialPort struct {
	steps       []serialReadStep
	readIndex   int
	closed      bool
	writtenData []byte
}

func (p *testSerialPort) Open(*serial.Config) error {
	return nil
}

func (p *testSerialPort) Read(dst []byte) (int, error) {
	if p.readIndex >= len(p.steps) {
		return 0, io.EOF
	}

	step := p.steps[p.readIndex]
	p.readIndex++

	n := copy(dst, step.data)
	if step.after != nil {
		step.after()
	}

	return n, step.err
}

func (p *testSerialPort) Write(src []byte) (int, error) {
	p.writtenData = append(p.writtenData, src...)
	return len(src), nil
}

func (p *testSerialPort) Close() error {
	p.closed = true
	return nil
}

func TestAcceptSerialRequests(t *testing.T) {
	t.Run("returns read error", func(t *testing.T) {
		s := NewServer()
		port := &testSerialPort{
			steps: []serialReadStep{{err: errors.New("read failure")}},
		}

		err := s.acceptSerialRequests(port)
		require.Error(t, err)
		assert.ErrorContains(t, err, "read failure")
		assert.True(t, port.closed)
	})

	t.Run("skips invalid frame and keeps running", func(t *testing.T) {
		s := NewServer()
		port := &testSerialPort{}
		port.steps = []serialReadStep{
			{data: []byte{0x01, 0x04, 0x02, 0xFF, 0xFF, 0xB8, 0x81}},
			{err: io.EOF, after: func() { close(s.closeSignalChan) }},
		}

		err := s.acceptSerialRequests(port)
		require.NoError(t, err)
		assert.Equal(t, 0, len(s.requestChan))
		assert.True(t, port.closed)
	})

	t.Run("queues valid request", func(t *testing.T) {
		s := NewServer()
		port := &testSerialPort{}
		port.steps = []serialReadStep{
			{data: []byte{0x01, 0x04, 0x02, 0xFF, 0xFF, 0xB8, 0x80}},
			{err: io.EOF, after: func() { close(s.closeSignalChan) }},
		}

		err := s.acceptSerialRequests(port)
		require.NoError(t, err)
		require.Len(t, s.requestChan, 1)

		request := <-s.requestChan
		require.NotNil(t, request)
		assert.EqualValues(t, 4, request.frame.GetFunction())
		assert.True(t, port.closed)
	})
}
