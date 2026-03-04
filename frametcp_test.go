package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTCPFrame(t *testing.T) {
	tests := []struct {
		name        string
		packet      []byte
		wantError   string
		wantFrame   *TCPFrame
		shouldError bool
	}{
		{
			name:   "valid packet",
			packet: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0xFF, 0x06, 0x01},
			wantFrame: &TCPFrame{
				TransactionIdentifier: 1,
				ProtocolIdentifier:    0,
				Length:                3,
				Device:                0xFF,
				Function:              0x06,
				Data:                  []byte{0x01},
			},
		},
		{
			name:        "packet too short",
			packet:      []byte{0x00, 0x01, 0x00, 0x00},
			shouldError: true,
			wantError:   "packet less than 9 bytes",
		},
		{
			name:        "invalid protocol id",
			packet:      []byte{0x00, 0x01, 0x00, 0x01, 0x00, 0x03, 0xFF, 0x06, 0x01},
			shouldError: true,
			wantError:   "invalid protocol identifier",
		},
		{
			name:        "length mismatch",
			packet:      []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0xFF, 0x06, 0x01},
			shouldError: true,
			wantError:   "specified packet length does not match actual packet length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := NewTCPFrame(tt.packet)
			if tt.shouldError {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantFrame, frame)
		})
	}
}

func TestTCPFrameMethods(t *testing.T) {
	frame := &TCPFrame{
		TransactionIdentifier: 2,
		ProtocolIdentifier:    0,
		Length:                4,
		Device:                0xFF,
		Function:              0x03,
		Data:                  []byte{0x00, 0x01},
	}

	t.Run("copy", func(t *testing.T) {
		copied := frame.Copy()
		require.IsType(t, &TCPFrame{}, copied)
		assert.Equal(t, frame.Bytes(), copied.Bytes())
	})

	t.Run("getters", func(t *testing.T) {
		assert.EqualValues(t, 0x03, frame.GetFunction())
		assert.Equal(t, []byte{0x00, 0x01}, frame.GetData())
	})

	t.Run("set data updates length", func(t *testing.T) {
		frame.SetData([]byte{0x10, 0x20, 0x30})
		assert.Equal(t, []byte{0x10, 0x20, 0x30}, frame.Data)
		assert.EqualValues(t, 5, frame.Length)
	})

	t.Run("set exception", func(t *testing.T) {
		frame.Function = 0x03
		frame.SetException(IllegalDataAddress)
		assert.EqualValues(t, 0x83, frame.Function)
		assert.Equal(t, []byte{byte(IllegalDataAddress)}, frame.Data)
		assert.EqualValues(t, 3, frame.Length)
	})

	t.Run("bytes", func(t *testing.T) {
		frame := &TCPFrame{
			TransactionIdentifier: 0x1234,
			ProtocolIdentifier:    0,
			Device:                0x01,
			Function:              0x04,
			Data:                  []byte{0xAA, 0xBB},
		}

		assert.Equal(t, []byte{0x12, 0x34, 0x00, 0x00, 0x00, 0x04, 0x01, 0x04, 0xAA, 0xBB}, frame.Bytes())
	})
}
