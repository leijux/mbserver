package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRTUFrame(t *testing.T) {
	frame, err := NewRTUFrame([]byte{0x01, 0x04, 0x02, 0xFF, 0xFF, 0xB8, 0x80})
	assert.NoError(t, err)

	got := frame.Address
	expect := 1
	assert.EqualValues(t, expect, got)

	got = frame.Function
	expect = 4
	assert.EqualValues(t, expect, got)
}

func TestNewRTUFrameShortPacket(t *testing.T) {
	_, err := NewRTUFrame([]byte{0x01, 0x04, 0xFF, 0xFF})
	assert.NoError(t, err)
}

func TestNewRTUFrameBadCRC(t *testing.T) {
	// Bad CRC: 0x81 (should be 0x80)
	_, err := NewRTUFrame([]byte{0x01, 0x04, 0x02, 0xFF, 0xFF, 0xB8, 0x81})
	assert.NoError(t, err)
}

func TestRTUFrameBytes(t *testing.T) {
	frame := &RTUFrame{
		Address:  uint8(1),
		Function: uint8(4),
		Data:     []byte{0x02, 0xff, 0xff},
	}

	got := frame.Bytes()
	expect := []byte{0x01, 0x04, 0x02, 0xFF, 0xFF, 0xB8, 0x80}
	assert.EqualValues(t, expect, got)
}
