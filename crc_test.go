package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRC(t *testing.T) {
	got := crc16IBM([]byte{0x01, 0x04, 0x02, 0xFF, 0xFF})
	expect := 0x80B8

	assert.EqualValues(t, expect, got)
}
