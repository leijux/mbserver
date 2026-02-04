package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemRegister_ReadCoils(t *testing.T) {
	mr := NewMemRegister()
	mr.Coils[0] = true
	mr.Coils[1] = false
	mr.Coils[2] = true

	tests := []struct {
		name      string
		start     int
		count     int
		wantData  []bool
		wantError Exception
	}{
		{"valid range", 0, 3, []bool{true, false, true}, Success},
		{"single coil", 1, 1, []bool{false}, Success},
		{"empty range", 5, 0, []bool{}, Success},
		{"out of bounds", 65534, 4, nil, IllegalDataAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, exc := mr.ReadCoils(tt.start, tt.count)
			require.Equal(t, tt.wantError, exc)
			assert.Equal(t, tt.wantData, data)
		})
	}
}

func TestMemRegister_ReadDiscreteInputs(t *testing.T) {
	mr := NewMemRegister()
	mr.DiscreteInputs[0] = true
	mr.DiscreteInputs[1] = false
	mr.DiscreteInputs[2] = true

	tests := []struct {
		name      string
		start     int
		count     int
		wantData  []bool
		wantError Exception
	}{
		{"valid range", 0, 3, []bool{true, false, true}, Success},
		{"single input", 1, 1, []bool{false}, Success},
		{"out of bounds", 65534, 4, nil, IllegalDataAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, exc := mr.ReadDiscreteInputs(tt.start, tt.count)
			require.Equal(t, tt.wantError, exc)
			assert.Equal(t, tt.wantData, data)
		})
	}
}

func TestMemRegister_ReadHoldingRegisters(t *testing.T) {
	mr := NewMemRegister()
	mr.HoldingRegisters[0] = 100
	mr.HoldingRegisters[1] = 200
	mr.HoldingRegisters[2] = 300

	tests := []struct {
		name      string
		start     int
		count     int
		wantData  []uint16
		wantError Exception
	}{
		{"valid range", 0, 3, []uint16{100, 200, 300}, Success},
		{"single register", 1, 1, []uint16{200}, Success},
		{"out of bounds", 65534, 4, nil, IllegalDataAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, exc := mr.ReadHoldingRegisters(tt.start, tt.count)
			require.Equal(t, tt.wantError, exc)
			assert.Equal(t, tt.wantData, data)
		})
	}
}

func TestMemRegister_ReadInputRegisters(t *testing.T) {
	mr := NewMemRegister()
	mr.InputRegisters[0] = 100
	mr.InputRegisters[1] = 200
	mr.InputRegisters[2] = 300

	tests := []struct {
		name      string
		start     int
		count     int
		wantData  []uint16
		wantError Exception
	}{
		{"valid range", 0, 3, []uint16{100, 200, 300}, Success},
		{"single register", 1, 1, []uint16{200}, Success},
		{"out of bounds", 65534, 4, nil, IllegalDataAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, exc := mr.ReadInputRegisters(tt.start, tt.count)
			require.Equal(t, tt.wantError, exc)
			assert.Equal(t, tt.wantData, data)
		})
	}
}

func TestMemRegister_WriteSingleCoil(t *testing.T) {
	mr := NewMemRegister()

	tests := []struct {
		name      string
		addr      int
		value     bool
		wantError Exception
		wantValue bool
	}{
		{"write true", 10, true, Success, true},
		{"write false", 10, false, Success, false},
		{"out of bounds", 65536, true, IllegalDataAddress, false},
		{"boundary", 65535, true, Success, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exc := mr.WriteSingleCoil(tt.addr, tt.value)
			require.Equal(t, tt.wantError, exc)
			if tt.wantError == Success {
				assert.Equal(t, tt.wantValue, mr.Coils[tt.addr])
			}
		})
	}
}

func TestMemRegister_WriteSingleRegister(t *testing.T) {
	mr := NewMemRegister()

	tests := []struct {
		name      string
		addr      int
		value     uint16
		wantError Exception
		wantValue uint16
	}{
		{"write value", 10, 12345, Success, 12345},
		{"write zero", 10, 0, Success, 0},
		{"out of bounds", 65536, 100, IllegalDataAddress, 0},
		{"boundary", 65535, 65535, Success, 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exc := mr.WriteSingleRegister(tt.addr, tt.value)
			require.Equal(t, tt.wantError, exc)
			if tt.wantError == Success {
				assert.Equal(t, tt.wantValue, mr.HoldingRegisters[tt.addr])
			}
		})
	}
}

func TestMemRegister_WriteMultipleCoils(t *testing.T) {
	mr := NewMemRegister()

	tests := []struct {
		name      string
		start     int
		values    []bool
		wantError Exception
		wantData  []bool
	}{
		{"write multiple", 10, []bool{true, false, true, true, false}, Success, []bool{true, false, true, true, false}},
		{"write single", 20, []bool{true}, Success, []bool{true}},
		{"write empty", 30, []bool{}, Success, []bool{}},
		{"out of bounds", 65535, []bool{true, false, true}, IllegalDataAddress, nil},
		{"partial out of bounds", 65535, []bool{true, true}, IllegalDataAddress, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exc := mr.WriteMultipleCoils(tt.start, tt.values)
			require.Equal(t, tt.wantError, exc)
			if tt.wantError == Success {
				assert.Equal(t, tt.wantData, mr.Coils[tt.start:tt.start+len(tt.values)])
			}
		})
	}
}

func TestMemRegister_WriteMultipleRegisters(t *testing.T) {
	mr := NewMemRegister()

	tests := []struct {
		name      string
		start     int
		values    []uint16
		wantError Exception
		wantData  []uint16
	}{
		{"write multiple", 10, []uint16{100, 200, 300, 400, 500}, Success, []uint16{100, 200, 300, 400, 500}},
		{"write single", 20, []uint16{12345}, Success, []uint16{12345}},
		{"write empty", 30, []uint16{}, Success, []uint16{}},
		{"out of bounds", 65535, []uint16{100, 200, 300}, IllegalDataAddress, nil},
		{"partial out of bounds", 65535, []uint16{100, 200}, IllegalDataAddress, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exc := mr.WriteMultipleRegisters(tt.start, tt.values)
			require.Equal(t, tt.wantError, exc)
			if tt.wantError == Success {
				assert.Equal(t, tt.wantData, mr.HoldingRegisters[tt.start:tt.start+len(tt.values)])
			}
		})
	}
}

func TestMemRegister_Init(t *testing.T) {
	mr := NewMemRegister()

	tests := []struct {
		name  string
		check func(t *testing.T, mr *MemRegister)
	}{
		{
			name: "all zeros",
			check: func(t *testing.T, mr *MemRegister) {
				assert.False(t, mr.Coils[0])
				assert.False(t, mr.Coils[65535])
				assert.False(t, mr.DiscreteInputs[0])
				assert.False(t, mr.DiscreteInputs[65535])
				assert.Equal(t, uint16(0), mr.HoldingRegisters[0])
				assert.Equal(t, uint16(0), mr.HoldingRegisters[65535])
				assert.Equal(t, uint16(0), mr.InputRegisters[0])
				assert.Equal(t, uint16(0), mr.InputRegisters[65535])
			},
		},
		{
			name: "default size",
			check: func(t *testing.T, mr *MemRegister) {
				assert.Len(t, mr.Coils, 65536)
				assert.Len(t, mr.DiscreteInputs, 65536)
				assert.Len(t, mr.HoldingRegisters, 65536)
				assert.Len(t, mr.InputRegisters, 65536)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, mr)
		})
	}
}
