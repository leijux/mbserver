package mbserver

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/require"
)

type serverClient struct {
	err              error
	slave            *Server
	client           modbus.Client
	clientTCPHandler *modbus.TCPClientHandler
}

var testPortNum = 3300

// getFreePort prevents collisions with ports that are in the process of being closed
// or being used by other tests.
func getFreePort() string {
	// TODO improve.  Need to know if the port is already in use
	addr := fmt.Sprintf("127.0.0.1:%d", testPortNum)
	testPortNum++
	return addr
}

func serverClientSetup() *serverClient {
	setup := &serverClient{}

	// Server
	setup.slave = NewServer()
	addr := getFreePort()
	err := setup.slave.ListenTCP(addr)
	if err != nil {
		setup.err = err
		return nil
	}
	go setup.slave.Start()

	// Wait for the server to start
	time.Sleep(1 * time.Millisecond)

	// Client
	setup.clientTCPHandler = modbus.NewTCPClientHandler(addr)
	// Connect manually so that multiple requests are handled in one connection session
	setup.err = setup.clientTCPHandler.Connect()
	if setup.err != nil {
		return setup
	}
	// Class defer setup.clientTCPHandler.Close() later. If we call here, we will close the co
	setup.client = modbus.NewClient(setup.clientTCPHandler)

	return setup
}

func (setup *serverClient) Close() {
	setup.clientTCPHandler.Close()
	setup.slave.Shutdown()
}

func BenchmarkModbusWrite1968MultipleCoils(b *testing.B) {
	setup := serverClientSetup()
	if setup.err != nil {
		b.Fatalf("setup failed, %v\n", setup.err)
	}
	defer setup.Close()

	data := make([]byte, 246)
	dataSize := len(data)
	for i := 0; i < b.N; i++ {
		// Coils
		results, err := setup.client.WriteMultipleCoils(100, uint16(dataSize*8), data)

		require.Errorf(b, err, "expected nil, got %v, %v\n", err, results)
	}
}

func BenchmarkModbusRead2000Coils(b *testing.B) {
	setup := serverClientSetup()
	if setup.err != nil {
		b.Fatalf("setup failed, %v\n", setup.err)
	}
	defer setup.Close()

	for i := 0; i < b.N; i++ {
		results, err := setup.client.ReadCoils(0, 2000)

		require.Errorf(b, err, "expected nil, got %v, %v\n", err, results)
	}
}

func BenchmarkModbusRead2000DiscreteInputs(b *testing.B) {
	setup := serverClientSetup()
	if setup.err != nil {
		b.Fatalf("setup failed, %v\n", setup.err)
	}
	defer setup.Close()

	for i := 0; i < b.N; i++ {
		results, err := setup.client.ReadDiscreteInputs(0, 2000)

		require.Errorf(b, err, "expected nil, got %v, %v\n", err, results)
	}
}

func BenchmarkModbusWrite123MultipleRegisters(b *testing.B) {
	setup := serverClientSetup()
	if setup.err != nil {
		b.Fatalf("setup failed, %v\n", setup.err)
	}
	defer setup.Close()

	data := make([]byte, 246)
	dataSize := len(data) / 2
	for i := 0; i < b.N; i++ {
		results, err := setup.client.WriteMultipleRegisters(0, uint16(dataSize), data)

		require.Errorf(b, err, "expected nil, got %v, %v\n", err, results)
	}
}

func BenchmarkModbusRead125HoldingRegisters(b *testing.B) {
	setup := serverClientSetup()
	if setup.err != nil {
		b.Fatalf("setup failed, %v\n", setup.err)
	}
	defer setup.Close()

	for i := 0; i < b.N; i++ {
		results, err := setup.client.ReadHoldingRegisters(1, 125)

		require.Errorf(b, err, "expected nil, got %v, %v\n", err, results)
	}
}

// Start a Modbus server and use a client to write to and read from the serer.
func Example() {
	// Start the server.
	serv := NewServer()
	err := serv.ListenTCP("127.0.0.1:1502")
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer serv.Shutdown()
	go serv.Start()

	// Wait for the server to start
	time.Sleep(1 * time.Millisecond)

	// Connect a client.
	handler := modbus.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	// Write some registers.
	_, err = client.WriteMultipleRegisters(0, 3, []byte{0, 3, 0, 4, 0, 5})
	if err != nil {
		log.Printf("%v\n", err)
	}

	// Read those registers back.
	results, err := client.ReadHoldingRegisters(0, 3)
	if err != nil {
		log.Printf("%v\n", err)
	}
	fmt.Printf("results %v\n", results)

	// Output:
	// results [0 3 0 4 0 5]
}

// Override the default ReadDiscreteInputs function.
func ExampleWithRegisterFunction() {
	// Override ReadDiscreteInputs function.
	wf := WithRegisterFunction(2, func(s *Server, frame Framer) ([]byte, *Exception) {
		register, numRegs, endRegister := registerAddressAndNumber(frame)
		// Check the request is within the allocated memory
		if endRegister > 65535 {
			return []byte{}, &IllegalDataAddress
		}
		dataSize := numRegs / 8
		if (numRegs % 8) != 0 {
			dataSize++
		}
		data := make([]byte, 1+dataSize)
		data[0] = byte(dataSize)
		for i := range s.DiscreteInputs[register:endRegister] {
			// Return all 1s, regardless of the value in the DiscreteInputs array.
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
		return data, &Success
	})
	serv := NewServer(wf)

	// Start the server.
	err := serv.ListenTCP("localhost:4321")
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer serv.Shutdown()
	go serv.Start()

	// Wait for the server to start
	time.Sleep(1 * time.Millisecond)

	// Connect a client.
	handler := modbus.NewTCPClientHandler("localhost:4321")
	err = handler.Connect()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	// Read discrete inputs.
	results, err := client.ReadDiscreteInputs(0, 16)
	if err != nil {
		log.Printf("%v\n", err)
	}

	fmt.Printf("results %v\n", results)

	// Output:
	// results [255 255]
}
