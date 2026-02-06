package mbserver

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

						header := make([]byte, 7)
						for i := 0; i < 7; {
							n, err := conn.Read(header[i:])
							if err != nil {
								var netErr net.Error
								if errors.As(err, &netErr) && netErr.Timeout() {
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
								return
							}
							i += n
						}

						pduLength := binary.BigEndian.Uint16(header[4:6])

						packet := make([]byte, 7+pduLength)
						copy(packet, header)
						remaining := packet[7:]

						for i := 0; i < int(pduLength); {
							n, err := conn.Read(remaining[i:])
							if err != nil {
								var netErr net.Error
								if errors.As(err, &netErr) && netErr.Timeout() {
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
								return
							}
							i += n
						}

						frame, err := NewTCPFrame(packet)
						if err != nil {
							slog.Error("failed to parse TCP frame", "error", err)

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
		return err
	}
	s.listeners = append(s.listeners, listen)
	return err
}

// ListenTLS starts the Modbus server listening on "address:port".
func (s *Server) ListenTLS(addressPort string, config *tls.Config) (err error) {
	listen, err := tls.Listen("tcp", addressPort, config)
	if err != nil {
		return err
	}
	s.listeners = append(s.listeners, listen)
	return err
}
