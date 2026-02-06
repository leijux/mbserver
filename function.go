package mbserver

import (
	"encoding/binary"
)

type Function func(Register, Framer) ([]byte, Exception)

// readCoils function 1, reads coils from internal memory.
func readCoils(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	coils, exception := r.ReadCoils(register, numRegs)
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
func readDiscreteInputs(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	discreteInputs, exception := r.ReadDiscreteInputs(register, numRegs)
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
func readHoldingRegisters(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	hRegisters, exception := r.ReadHoldingRegisters(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	data := make([]byte, 1, 1+numRegs*2)
	data[0] = byte(numRegs * 2)
	data = append(data, Uint16ToBytes(hRegisters)...)

	return data, Success
}

// readInputRegisters function 4, reads input registers from internal memory.
func readInputRegisters(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	iRegisters, exception := r.ReadInputRegisters(register, numRegs)
	if exception != Success {
		return []byte{}, exception
	}

	data := make([]byte, 1, 1+numRegs*2)
	data[0] = byte(numRegs * 2)
	data = append(data, Uint16ToBytes(iRegisters)...)

	return data, Success
}

// writeSingleCoil function 5, write a coil to internal memory.
func writeSingleCoil(r Register, frame Framer) ([]byte, Exception) {
	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1 // Modbus standard uses 0 for off and 1 for on
	}

	if exception := r.WriteSingleCoil(register, value != 0); exception != Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], Success
}

// writeSingleRegister function 6, write a holding register to internal memory.
func writeSingleRegister(r Register, frame Framer) ([]byte, Exception) {
	register, value := registerAddressAndValue(frame)

	if exception := r.WriteSingleRegister(register, value); exception != Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], Success
}

// writeMultipleCoils function 15, writes holding registers to internal memory.
func writeMultipleCoils(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	expectedBytes := (numRegs + 7) / 8
	if len(valueBytes) < expectedBytes {
		return []byte{}, IllegalDataValue
	}

	bitCount := 0
	bitValue := make([]bool, numRegs)

	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			bitValue[(i*8)+int(bitPos)] = bitAtPosition(value, bitPos) != 0
			bitCount++
			if bitCount >= numRegs {
				break
			}
		}
		if bitCount >= numRegs {
			break
		}
	}

	if exception := r.WriteMultipleCoils(register, bitValue); exception != Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], Success
}

// writeMultipleRegisters function 16, writes holding registers to internal memory.
func writeMultipleRegisters(r Register, frame Framer) ([]byte, Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register > 65535 || numRegs > 65535-register {
		return []byte{}, IllegalDataAddress
	}

	if len(valueBytes)/2 != numRegs {
		return []byte{}, IllegalDataValue
	}

	values := BytesToUint16(valueBytes)
	if exception := r.WriteMultipleRegisters(register, values); exception != Success {
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
