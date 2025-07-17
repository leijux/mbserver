package mbserver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Function 1
func TestReadCoils(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	// Set the coil values
	mr.Coils[10] = true
	mr.Coils[11] = true
	mr.Coils[17] = true
	mr.Coils[18] = true

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 1
	SetDataWithRegisterAndNumber(&frame, 10, 9)

	var req Request
	req.frame = &frame
	response := s.handle(&req)

	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	// 2 bytes, 0b1000011, 0b00000001
	expect := []byte{2, 131, 1}
	got := response.GetData()

	assert.Equal(t, expect, got)
}

// Function 2
func TestReadDiscreteInputs(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	// Set the discrete input values
	mr.DiscreteInputs[0] = true
	mr.DiscreteInputs[7] = true
	mr.DiscreteInputs[8] = true
	mr.DiscreteInputs[9] = true

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 2
	SetDataWithRegisterAndNumber(&frame, 0, 10)

	var req Request
	req.frame = &frame
	response := s.handle(&req)

	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := []byte{2, 129, 3}
	got := response.GetData()

	assert.Equal(t, expect, got)
}

// Function 3
func TestReadHoldingRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.HoldingRegisters[100] = 1
	mr.HoldingRegisters[101] = 2
	mr.HoldingRegisters[102] = 65535

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 3
	SetDataWithRegisterAndNumber(&frame, 100, 3)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := []byte{6, 0, 1, 0, 2, 255, 255}
	got := response.GetData()
	assert.Equal(t, expect, got)
}

// Function 4
func TestReadInputRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))
	mr.InputRegisters[200] = 1
	mr.InputRegisters[201] = 2
	mr.InputRegisters[202] = 65535

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255
	frame.Function = 4
	SetDataWithRegisterAndNumber(&frame, 200, 3)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := []byte{6, 0, 1, 0, 2, 255, 255}
	got := response.GetData()
	assert.Equal(t, expect, got)
}

// Function 5
func TestWriteSingleCoil(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 5
	SetDataWithRegisterAndNumber(&frame, 65535, 1024)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := true
	got := mr.Coils[65535]
	assert.Equal(t, expect, got)
}

// Function 6
func TestWriteHoldingRegister(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 6
	SetDataWithRegisterAndNumber(&frame, 5, 6)

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := uint16(6)
	got := mr.HoldingRegisters[5]
	assert.Equal(t, expect, got)
}

// Function 15
func TestWriteMultipleCoils(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 15
	SetDataWithRegisterAndNumberAndBytes(&frame, 1, 2, []byte{3})

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := []bool{true, true}
	got := mr.Coils[1:3]
	assert.Equal(t, expect, got)
}

// Function 16
func TestWriteHoldingRegisters(t *testing.T) {
	mr := NewMemRegister()
	s := NewServer(WithRegister(mr))

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 12
	frame.Device = 255
	frame.Function = 16
	SetDataWithRegisterAndNumberAndValues(&frame, 1, 2, []uint16{3, 4})

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	require.ErrorIs(t, exception, Success)

	expect := []uint16{3, 4}
	got := mr.HoldingRegisters[1:3]
	assert.Equal(t, expect, got)
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

	var frame TCPFrame
	frame.TransactionIdentifier = 1
	frame.ProtocolIdentifier = 0
	frame.Length = 6
	frame.Device = 255

	var req Request
	req.frame = &frame

	// bits
	SetDataWithRegisterAndNumber(&frame, 65535, 2)

	frame.Function = 1
	response := s.handle(&req)
	exception := GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	frame.Function = 2
	response = s.handle(&req)
	exception = GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	SetDataWithRegisterAndNumberAndBytes(&frame, 65535, 2, []byte{3})
	frame.Function = 15
	response = s.handle(&req)
	exception = GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	// registers
	SetDataWithRegisterAndNumber(&frame, 65535, 2)

	frame.Function = 3
	response = s.handle(&req)
	exception = GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	frame.Function = 4
	response = s.handle(&req)
	exception = GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}

	SetDataWithRegisterAndNumberAndValues(&frame, 65535, 2, []uint16{0, 0})
	frame.Function = 16
	response = s.handle(&req)
	exception = GetException(response)
	if !errors.Is(exception, IllegalDataAddress) {
		t.Errorf("expected IllegalDataAddress, got %v", exception.String())
	}
}
