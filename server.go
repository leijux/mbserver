// Package mbserver implements a Modbus server (slave).
package mbserver

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/goburrow/serial"
)

// Server is a Modbus slave with allocated memory for discrete inputs, coils, etc.
type Server struct {
	listeners []net.Listener
	ports     []serial.Port

	wg              sync.WaitGroup
	closeSignalChan chan struct{}

	requestChan chan *Request

	function         [256]function
	DiscreteInputs   []byte
	Coils            []byte
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

// Request contains the connection and Modbus frame.
type Request struct {
	conn  io.ReadWriteCloser
	frame Framer
}

// OptionFunc is a function type used to configure options for the Server.
type OptionFunc func(s *Server)

// WithDiscreteInputs sets the DiscreteInputs data for the Server.
// Parameter data is a byte slice containing the discrete input data to set.
func WithDiscreteInputs(data []byte) OptionFunc {
	return func(s *Server) {
		s.DiscreteInputs = data
	}
}

// WithCoils sets the Coils data for the Server.
// Parameter data is a byte slice containing the coil data to set.
func WithCoils(data []byte) OptionFunc {
	return func(s *Server) {
		s.Coils = data
	}
}

// WithHoldingRegisters sets the HoldingRegisters data for the Server.
// Parameter data is a byte slice containing the holding register data to set.
func WithHoldingRegisters(data []uint16) OptionFunc {
	return func(s *Server) {
		s.HoldingRegisters = data
	}
}

// WithInputRegisters sets the InputRegisters data for the Server.
// Parameter data is a byte slice containing the input register data to set.
func WithInputRegisters(data []uint16) OptionFunc {
	return func(s *Server) {
		s.InputRegisters = data
	}
}

// WithRegisterFunction registers a custom function handler for a specific function code.
// Parameter funcCode is the function code, and function is the custom handler for that code.
func WithRegisterFunction(funcCode uint8, function function) OptionFunc {
	return func(s *Server) {
		s.registerFunctionHandler(funcCode, function)
	}
}

// NewServer creates a new Modbus server (slave).
func NewServer(opts ...OptionFunc) *Server {
	s := &Server{}

	// Add default functions.
	s.function[1] = ReadCoils
	s.function[2] = ReadDiscreteInputs
	s.function[3] = ReadHoldingRegisters
	s.function[4] = ReadInputRegisters
	s.function[5] = WriteSingleCoil
	s.function[6] = WriteHoldingRegister
	s.function[15] = WriteMultipleCoils
	s.function[16] = WriteHoldingRegisters

	for _, opt := range opts {
		opt(s)
	}

	// Allocate Modbus memory maps.
	if s.DiscreteInputs == nil {
		s.DiscreteInputs = make([]byte, 65536)
	}

	if s.Coils == nil {
		s.Coils = make([]byte, 65536)
	}

	if s.HoldingRegisters == nil {
		s.HoldingRegisters = make([]uint16, 65536)
	}

	if s.InputRegisters == nil {
		s.InputRegisters = make([]uint16, 65536)
	}

	s.requestChan = make(chan *Request)
	s.closeSignalChan = make(chan struct{})

	return s
}

// registerFunctionHandler override the default behavior for a given Modbus function.
func (s *Server) registerFunctionHandler(funcCode uint8, function function) {
	s.function[funcCode] = function
}

func (s *Server) handle(request *Request) Framer {
	var exception *Exception
	var data []byte

	response := request.frame.Copy()

	funcCode := request.frame.GetFunction()
	if s.function[funcCode] != nil {
		data, exception = s.function[funcCode](s, request.frame)
		response.SetData(data)
	} else {
		exception = &IllegalFunction
	}

	if !errors.Is(*exception, Success) {
		response.SetException(exception)
	}

	return response
}

// All requests are handled synchronously to prevent modbus memory corruption.
func (s *Server) handler() {
	for {
		select {
		case <-s.closeSignalChan:
			return
		case request := <-s.requestChan:
			response := s.handle(request)
			request.conn.Write(response.Bytes())
		}
	}
}

// Start the service
func (s *Server) Start() {
	for _, listener := range s.listeners {
		go s.accept(listener)
	}

	for _, port := range s.ports {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.acceptSerialRequests(port)
		}()
	}

	s.handler()
}

// Shutdown stops listening to TCP/IP ports and closes serial ports.
func (s *Server) Shutdown() {
	close(s.closeSignalChan)

	s.wg.Wait()
}
