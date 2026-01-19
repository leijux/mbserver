package mbserver

type Register interface {
	ReadCoils(int, int) ([]bool, Exception)
	ReadDiscreteInputs(int, int) ([]bool, Exception)
	ReadHoldingRegisters(int, int) ([]uint16, Exception)
	ReadInputRegisters(int, int) ([]uint16, Exception)

	WriteSingleCoil(int, bool) Exception
	WriteSingleRegister(int, uint16) Exception
	WriteMultipleCoils(int, []bool) Exception
	WriteMultipleRegisters(int, []uint16) Exception
}

type MemRegister struct {
	Coils          []bool
	DiscreteInputs []bool

	HoldingRegisters []uint16
	InputRegisters   []uint16
}

var _ Register = (*MemRegister)(nil)

func NewMemRegister() *MemRegister {
	return &MemRegister{
		Coils:            make([]bool, 65536),
		DiscreteInputs:   make([]bool, 65536),
		HoldingRegisters: make([]uint16, 65536),
		InputRegisters:   make([]uint16, 65536),
	}
}

func (r *MemRegister) ReadCoils(start, count int) ([]bool, Exception) {
	if start+count > len(r.Coils) {
		return nil, IllegalDataAddress
	}
	return r.Coils[start : start+count], Success
}

func (r *MemRegister) ReadDiscreteInputs(start, count int) ([]bool, Exception) {
	if start+count > len(r.DiscreteInputs) {
		return nil, IllegalDataAddress
	}
	return r.DiscreteInputs[start : start+count], Success
}

func (r *MemRegister) ReadHoldingRegisters(start, count int) ([]uint16, Exception) {
	if start+count > len(r.HoldingRegisters) {
		return nil, IllegalDataAddress
	}
	return r.HoldingRegisters[start : start+count], Success
}

func (r *MemRegister) ReadInputRegisters(start, count int) ([]uint16, Exception) {
	if start+count > len(r.InputRegisters) {
		return nil, IllegalDataAddress
	}
	return r.InputRegisters[start : start+count], Success
}

func (r *MemRegister) WriteSingleCoil(start int, value bool) Exception {
	if start >= len(r.Coils) {
		return IllegalDataAddress
	}
	r.Coils[start] = value
	return Success
}

func (r *MemRegister) WriteSingleRegister(start int, value uint16) Exception {
	if start >= len(r.HoldingRegisters) {
		return IllegalDataAddress
	}
	r.HoldingRegisters[start] = value
	return Success
}

func (r *MemRegister) WriteMultipleCoils(start int, values []bool) Exception {
	if start+len(values) > len(r.Coils) {
		return IllegalDataAddress
	}
	for i, value := range values {
		r.Coils[start+i] = value
	}
	return Success
}

func (r *MemRegister) WriteMultipleRegisters(start int, values []uint16) Exception {
	if start+len(values) > len(r.HoldingRegisters) {
		return IllegalDataAddress
	}
	for i, value := range values {
		r.HoldingRegisters[start+i] = value
	}
	return Success
}
