//go:build linux
// +build linux

package mbserver

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The serial read and close has a known race condition.
// https://github.com/golang/go/issues/10001
func TestModbusRTU(t *testing.T) {
	if _, err := exec.LookPath("socat"); err != nil {
		t.Skip("socat not found in PATH")
	}

	ttyDir := t.TempDir()
	serverTTY := filepath.Join(ttyDir, "ttyFOO")
	clientTTY := filepath.Join(ttyDir, "ttyBAR")

	// Create a pair of virtual serial devices.
	cmd := exec.Command("socat",
		"pty,raw,echo=0,link="+serverTTY,
		"pty,raw,echo=0,link="+clientTTY)
	err := cmd.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	// Wait for the virtual serial devices to be created.
	require.Eventually(t, func() bool {
		_, fooErr := os.Stat(serverTTY)
		_, barErr := os.Stat(clientTTY)
		return fooErr == nil && barErr == nil
	}, time.Second, 10*time.Millisecond)

	// Server
	s := NewServer()
	err = s.ListenRTU(&serial.Config{
		Address:  serverTTY,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second})
	require.NoError(t, err)

	t.Cleanup(s.Shutdown)
	go s.Start()

	// Allow the server to start and avoid connection refused on the client.
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewRTUClientHandler(clientTTY)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	require.NoError(t, err)

	t.Cleanup(func() { _ = handler.Close() })
	client := modbus.NewClient(handler)

	// Coils
	_, err = client.WriteMultipleCoils(100, 9, []byte{255, 1})
	require.NoError(t, err)

	results, err := client.ReadCoils(100, 16)
	require.NoError(t, err)

	assert.EqualValues(t, []byte{255, 1}, results)
}
