package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTCPFrame(function uint8) *TCPFrame {
	return &TCPFrame{
		TransactionIdentifier: 1,
		ProtocolIdentifier:    0,
		Length:                6,
		Device:                255,
		Function:              function,
	}
}

func assertSuccess(t *testing.T, response Framer) {
	t.Helper()
	require.Equal(t, Success, GetException(response))
}

// Function 1
func TestReadCoils(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.Coils[10] = true
	mr.Coils[11] = true
	mr.Coils[17] = true
	mr.Coils[18] = true

	frame := newTestTCPFrame(1)
	SetDataWithRegisterAndNumber(frame, 10, 9)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)

	expected := []byte{2, 131, 1}
	assert.Equal(t, expected, response.GetData())
}

// Function 2
func TestReadDiscreteInputs(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.DiscreteInputs[0] = true
	mr.DiscreteInputs[7] = true
	mr.DiscreteInputs[8] = true
	mr.DiscreteInputs[9] = true

	frame := newTestTCPFrame(2)
	SetDataWithRegisterAndNumber(frame, 0, 10)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)

	expected := []byte{2, 129, 3}
	assert.Equal(t, expected, response.GetData())
}

// Function 3
func TestReadHoldingRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.HoldingRegisters[100] = 1
	mr.HoldingRegisters[101] = 2
	mr.HoldingRegisters[102] = 65535

	frame := newTestTCPFrame(3)
	SetDataWithRegisterAndNumber(frame, 100, 3)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)

	expected := []byte{6, 0, 1, 0, 2, 255, 255}
	assert.Equal(t, expected, response.GetData())
}

// Function 4
func TestReadInputRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.InputRegisters[200] = 1
	mr.InputRegisters[201] = 2
	mr.InputRegisters[202] = 65535

	frame := newTestTCPFrame(4)
	SetDataWithRegisterAndNumber(frame, 200, 3)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)

	expected := []byte{6, 0, 1, 0, 2, 255, 255}
	assert.Equal(t, expected, response.GetData())
}

// Function 5
func TestWriteSingleCoil(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	frame := newTestTCPFrame(5)
	SetDataWithRegisterAndNumber(frame, 65535, 1024)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)
	assert.Equal(t, true, mr.Coils[65535])
}

// Function 6
func TestWriteHoldingRegister(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	frame := newTestTCPFrame(6)
	SetDataWithRegisterAndNumber(frame, 5, 6)

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)
	assert.Equal(t, uint16(6), mr.HoldingRegisters[5])
}

// Function 15
func TestWriteMultipleCoils(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	frame := newTestTCPFrame(15)
	SetDataWithRegisterAndNumberAndBytes(frame, 1, 2, []byte{3})

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)
	assert.Equal(t, []bool{true, true}, mr.Coils[1:3])
}

// Function 16
func TestWriteHoldingRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	frame := newTestTCPFrame(16)
	SetDataWithRegisterAndNumberAndValues(frame, 1, 2, []uint16{3, 4})

	response := s.handle(&Request{frame: frame})
	assertSuccess(t, response)
	assert.Equal(t, []uint16{3, 4}, mr.HoldingRegisters[1:3])
}

func TestBytesToUint16(t *testing.T) {
	bytes := []byte{1, 2, 3, 4}
	got := BytesToUint16(bytes)
	expect := []uint16{258, 772}
	assert.Equal(t, expect, got)
}

func TestUint16ToBytes(t *testing.T) {
	values := []uint16{1, 2, 3}
	got := Uint16ToBytes(values)
	expect := []byte{0, 1, 0, 2, 0, 3}
	assert.Equal(t, expect, got)
}

func TestOutOfBounds(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	tests := []struct {
		name     string
		function uint8
		setData  func(*TCPFrame)
		expect   Exception
	}{
		{
			name:     "read coils overflow",
			function: 1,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumber(frame, 65535, 2)
			},
			expect: IllegalDataAddress,
		},
		{
			name:     "read discrete inputs overflow",
			function: 2,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumber(frame, 65535, 2)
			},
			expect: IllegalDataAddress,
		},
		{
			name:     "write multiple coils overflow",
			function: 15,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumberAndBytes(frame, 65535, 2, []byte{3})
			},
			expect: IllegalDataAddress,
		},
		{
			name:     "read holding registers overflow",
			function: 3,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumber(frame, 65535, 2)
			},
			expect: IllegalDataAddress,
		},
		{
			name:     "read input registers overflow",
			function: 4,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumber(frame, 65535, 2)
			},
			expect: IllegalDataAddress,
		},
		{
			name:     "write holding registers overflow",
			function: 16,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumberAndValues(frame, 65535, 2, []uint16{0, 0})
			},
			expect: IllegalDataAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := newTestTCPFrame(tt.function)
			tt.setData(frame)

			response := s.handle(&Request{frame: frame})
			require.Equal(t, tt.expect, GetException(response))
		})
	}
}

func TestIllegalDataValue(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	tests := []struct {
		name     string
		function uint8
		setData  func(*TCPFrame)
	}{
		{
			name:     "write multiple coils missing payload bytes",
			function: 15,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumberAndBytes(frame, 1, 10, []byte{1})
			},
		},
		{
			name:     "write holding registers mismatched payload",
			function: 16,
			setData: func(frame *TCPFrame) {
				SetDataWithRegisterAndNumberAndBytes(frame, 1, 2, []byte{0, 1})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := newTestTCPFrame(tt.function)
			tt.setData(frame)

			response := s.handle(&Request{frame: frame})
			require.Equal(t, IllegalDataValue, GetException(response))
		})
	}
}
