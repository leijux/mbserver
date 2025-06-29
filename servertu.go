package mbserver

import (
	"fmt"
	"io"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)

	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) error {
	defer port.Close()

SkipFrameError:
	for {
		select {
		case <-s.closeSignalChan:
			return nil
		default:
			buffer := make([]byte, 512)

			bytesRead, err := port.Read(buffer)
			if err != nil && err != io.EOF {
				return err
			}

			if bytesRead != 0 {

				// Set the length of the packet to the number of read bytes.
				packet := buffer[:bytesRead]

				frame, err := NewRTUFrame(packet)
				if err != nil {
					// s.l.Error("bad serial frame error", "err", err)
					//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous
					//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
					// s.l.Warn("Keep the RTU server running!!")
					continue SkipFrameError
					//return
				}

				request := &Request{port, frame}

				select {
				case s.requestChan <- request:
				case <-s.closeSignalChan:
					return nil
				}
			}
		}
	}
}
