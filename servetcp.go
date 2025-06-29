package mbserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

func (s *Server) accept(listen net.Listener) error {
	defer listen.Close()

	for {
		select {
		case <-s.closeSignalChan:
			return nil
		default:
			conn, err := listen.Accept()
			if err != nil {
				var ne net.Error
				if errors.As(err, &ne) && ne.Timeout() {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				select {
				case <-s.closeSignalChan:
					return nil
				default:
					return fmt.Errorf("unable to accept connections: %w", err)
				}
			}

			s.wg.Add(1)

			go func(conn net.Conn) {
				defer s.wg.Done()
				defer conn.Close()

				for {
					select {
					case <-s.closeSignalChan:
						return
					default:
						conn.SetReadDeadline(time.Now().Add(10 * time.Second))

						packet := make([]byte, 512)
						bytesRead, err := conn.Read(packet)
						if err != nil {
							var netErr net.Error
							if errors.As(err, &netErr) && netErr.Timeout() {
								// 超时，检查是否需要关闭
								select {
								case <-s.closeSignalChan:
									return
								default:
									continue
								}
							}

							if err != io.EOF {
								return
							}
						}
						// Set the length of the packet to the number of read bytes.
						packet = packet[:bytesRead]

						frame, err := NewTCPFrame(packet)
						if err != nil {
							return
						}

						request := &Request{conn, frame}

						select {
						case s.requestChan <- request:
						case <-s.closeSignalChan:
							return
						}
					}
				}
			}(conn)
		}

	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addressPort, err)
	}
	s.listeners = append(s.listeners, listen)
	return err
}

// ListenTLS starts the Modbus server listening on "address:port".
func (s *Server) ListenTLS(addressPort string, config *tls.Config) (err error) {
	listen, err := tls.Listen("tcp", addressPort, config)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addressPort, err)
	}
	s.listeners = append(s.listeners, listen)
	return err
}
