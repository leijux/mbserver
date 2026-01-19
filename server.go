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

	function [256]Function

	register Register
}

// Request contains the connection and Modbus frame.
type Request struct {
	conn  io.ReadWriteCloser
	frame Framer
}

// OptionFunc is a function type used to configure options for the Server.
type OptionFunc func(s *Server)

// WithRegisterFunction registers a custom function handler for a specific function code.
// Parameter funcCode is the function code, and function is the custom handler for that code.
func WithRegisterFunction(funcCode uint8, function Function) OptionFunc {
	return func(s *Server) {
		s.registerFunctionHandler(funcCode, function)
	}
}

// WithRegister sets the memory register for the server.
func WithRegister(register Register) OptionFunc {
	return func(s *Server) {
		s.register = register
	}
}

// NewServer creates a new Modbus server (slave).
func NewServer(opts ...OptionFunc) *Server {
	s := &Server{}

	// Add default functions.
	s.function[1] = readCoils
	s.function[2] = readDiscreteInputs
	s.function[3] = readHoldingRegisters
	s.function[4] = readInputRegisters
	s.function[5] = writeSingleCoil
	s.function[6] = writeSingleRegister
	s.function[15] = writeMultipleCoils
	s.function[16] = writeMultipleRegisters

	for _, opt := range opts {
		opt(s)
	}

	if s.register == nil {
		s.register = &MemRegister{
			Coils:            make([]bool, 65536),
			DiscreteInputs:   make([]bool, 65536),
			HoldingRegisters: make([]uint16, 65536),
			InputRegisters:   make([]uint16, 65536),
		}
	}

	s.requestChan = make(chan *Request)
	s.closeSignalChan = make(chan struct{})

	return s
}

// registerFunctionHandler override the default behavior for a given Modbus function.
func (s *Server) registerFunctionHandler(funcCode uint8, f Function) {
	s.function[funcCode] = f
}

func (s *Server) handle(request *Request) Framer {
	var (
		exception Exception
		data      []byte
	)

	response := request.frame.Copy()

	funcCode := request.frame.GetFunction()
	if s.function[funcCode] != nil {
		data, exception = s.function[funcCode](s, request.frame)
		response.SetData(data)
	} else {
		exception = IllegalFunction
	}

	if !errors.Is(exception, Success) {
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

	//close the listeners
	for _, listener := range s.listeners {
		listener.Close()
	}

	//close the ports
	for _, port := range s.ports {
		port.Close()
	}
}
