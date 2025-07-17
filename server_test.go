package mbserver

import (
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAduRegisterAndNumber(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumber(&frame, 0, 64)

	expect := []byte{0, 0, 0, 64}
	got := frame.Data
	assert.Equal(t, expect, got)
}

func TestAduSetDataWithRegisterAndNumberAndValues(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumberAndValues(&frame, 7, 2, []uint16{3, 4})

	expect := []byte{0, 7, 0, 2, 4, 0, 3, 0, 4}
	got := frame.Data
	assert.Equal(t, expect, got)
}

func TestUnsupportedFunction(t *testing.T) {
	s := NewServer()
	var frame TCPFrame
	frame.Function = 255

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)

	assert.Equalf(t, exception, IllegalFunction, "expected IllegalFunction (%d), got (%v)", IllegalFunction, exception)
}

func TestModbus(t *testing.T) {
	mr := NewMemRegister()
	// Server
	s := NewServer(WithRegister(mr))
	err := s.ListenTCP("127.0.0.1:3333")
	require.NoError(t, err)
	t.Cleanup(s.Shutdown)
	go s.Start()

	// Allow the server to start and to avoid a connection refused on the client
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewTCPClientHandler("127.0.0.1:3333")
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	require.NoError(t, err)

	t.Cleanup(func() { handler.Close() })

	client := modbus.NewClient(handler)

	t.Run("Coils", func(t *testing.T) {
		results, err := client.WriteMultipleCoils(100, 9, []byte{255, 1})
		require.NoError(t, err)

		results, err = client.ReadCoils(100, 16)
		require.NoError(t, err)

		expect := []byte{255, 1}
		got := results
		assert.Equal(t, expect, got)
	})

	// Discrete inputs
	t.Run("Discrete inputs", func(t *testing.T) {
		results, err := client.ReadDiscreteInputs(0, 64)
		require.NoError(t, err)

		// test: 2017/05/14 21:09:53 modbus: sending 00 01 00 00 00 06 ff 02 00 00 00 40
		// test: 2017/05/14 21:09:53 modbus: received 00 01 00 00 00 0b ff 02 08 00 00 00 00 00 00 00 00
		expect := []byte{0, 0, 0, 0, 0, 0, 0, 0}
		got := results
		assert.Equal(t, expect, got)
	})

	t.Run("Holding registers", func(t *testing.T) {
		// Holding registers
		results, err := client.WriteMultipleRegisters(1, 2, []byte{0, 3, 0, 4})
		require.NoError(t, err)

		// received: 00 01 00 00 00 06 ff 10 00 01 00 02
		expect := []byte{0, 2}
		got := results
		assert.Equal(t, expect, got)

		results, err = client.ReadHoldingRegisters(1, 2)
		require.NoError(t, err)

		expect = []byte{0, 3, 0, 4}
		got = results
		assert.Equal(t, expect, got)
	})

	t.Run("Input registers", func(t *testing.T) {
		// Input registers
		mr.InputRegisters[65530] = 1
		mr.InputRegisters[65535] = 65535

		results, err := client.ReadInputRegisters(65530, 6)
		require.NoError(t, err)

		expect := []byte{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
		got := results
		assert.Equal(t, expect, got)
	})
}
