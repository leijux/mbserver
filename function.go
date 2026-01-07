package mbserver

import (
	"encoding/binary"
)

type Function func(*Server, Framer) ([]byte, Exception)

// readCoils function 1, reads coils from internal memory.
func readCoils(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register+numRegs > 65536 {
		return []byte{}, IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	coils, exception := s.register.ReadCoils(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	for i, value := range coils {
		if value {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, Success
}

// readDiscreteInputs function 2, reads discrete inputs from internal memory.
func readDiscreteInputs(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register+numRegs > 65536 {
		return []byte{}, IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	discreteInputs, exception := s.register.ReadDiscreteInputs(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	for i, value := range discreteInputs {
		if value {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, Success

}

// readHoldingRegisters function 3, reads holding registers from internal memory.
func readHoldingRegisters(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register+numRegs > 65536 {
		return []byte{}, IllegalDataAddress
	}

	hRegisters, exception := s.register.ReadHoldingRegisters(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	return append([]byte{byte(numRegs * 2)}, Uint16ToBytes(hRegisters)...), Success
}

// readInputRegisters function 4, reads input registers from internal memory.
func readInputRegisters(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register+numRegs > 65536 {
		return []byte{}, IllegalDataAddress
	}

	iRegisters, exception := s.register.ReadInputRegisters(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	return append([]byte{byte(numRegs * 2)}, Uint16ToBytes(iRegisters)...), Success
}

// writeSingleCoil function 5, write a coil to internal memory.
func writeSingleCoil(s *Server, frame Framer) ([]byte, Exception) {
	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1 // Modbus standard uses 0 for off and 1 for on
	}
	s.register.WriteCoils(register, value != 0)
	return frame.GetData()[0:4], Success
}

// writeHoldingRegister function 6, write a holding register to internal memory.
func writeHoldingRegister(s *Server, frame Framer) ([]byte, Exception) {
	register, value := registerAddressAndValue(frame)

	s.register.WriteHoldingRegister(register, value)
	return frame.GetData()[0:4], Success
}

// writeMultipleCoils function 15, writes holding registers to internal memory.
func writeMultipleCoils(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register+numRegs > 65535 {
		return []byte{}, IllegalDataAddress
	}

	// TODO This is not correct, bits and bytes do not always align
	//if len(valueBytes)/2 != numRegs {
	//	return []byte{}, &IllegalDataAddress
	//}

	bitCount := 0
	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			s.register.WriteCoils(register+(i*8)+int(bitPos), bitAtPosition(value, bitPos) != 0)
			bitCount++
			if bitCount >= numRegs {
				break
			}
		}
		if bitCount >= numRegs {
			break
		}
	}

	return frame.GetData()[0:4], Success
}

// writeHoldingRegisters function 16, writes holding registers to internal memory.
func writeHoldingRegisters(s *Server, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register+numRegs > 65535 {
		return []byte{}, IllegalDataAddress
	}

	if len(valueBytes)/2 != numRegs {
		return []byte{}, IllegalDataAddress
	}

	values := BytesToUint16(valueBytes)
	if exception := s.register.WriteHoldingRegisters(register, values); exception != Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], Success
}

// BytesToUint16 converts a big endian array of bytes to an array of unit16s
func BytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

// Uint16ToBytes converts an array of uint16s to a big endian array of bytes
func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

func bitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}
