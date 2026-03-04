package mbserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExceptionError(t *testing.T) {
	tests := []struct {
		name      string
		exception Exception
		expected  string
	}{
		{
			name:      "known exception",
			exception: IllegalDataValue,
			expected:  "IllegalDataValue",
		},
		{
			name:      "unknown exception",
			exception: Exception(9),
			expected:  "Exception(9)",
		},
		{
			name:      "gateway exception",
			exception: GatewayTargetDeviceFailedToRespond,
			expected:  "GatewayTargetDeviceFailedToRespond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.exception.Error())
		})
	}
}
