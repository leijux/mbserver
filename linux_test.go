//go:build linux
// +build linux

package mbserver

import (
	"os/exec"
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
	// Create a pair of virutal serial devices.
	cmd := exec.Command("socat",
		"pty,raw,echo=0,link=ttyFOO",
		"pty,raw,echo=0,link=ttyBAR")
	err := cmd.Start()
	require.NoError(t, err)

	defer cmd.Wait()
	defer cmd.Process.Kill()

	// Allow the virutal serial devices to be created.
	time.Sleep(10 * time.Millisecond)

	// Server
	s := NewServer()
	err = s.ListenRTU(&serial.Config{
		Address:  "ttyFOO",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second})
	require.NoError(t, err)

	defer s.Shutdown()
	go s.Start()

	// Allow the server to start and to avoid a connection refused on the client
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewRTUClientHandler("ttyBAR")
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	require.NoError(t, err)

	defer handler.Close()
	client := modbus.NewClient(handler)

	// Coils
	_, err = client.WriteMultipleCoils(100, 9, []byte{255, 1})
	require.NoError(t, err)

	results, err := client.ReadCoils(100, 16)
	require.NoError(t, err)

	expect := []byte{255, 1}
	got := results
	assert.EqualValues(t, expect, got)
}
