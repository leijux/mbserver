package mbserver

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		select {
		case <-s.closeSignalChan:
			return listen.Close()
		default:
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
				s.l.Error("Unable to accept connections", "err", err)
			return err
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
				packet := make([]byte, 512)
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
								s.l.Error("read eroor", "err", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
							s.l.Error("bad packet error", "err", err)
					return
				}

				request := &Request{conn, frame}

				s.requestChan <- request
				}

			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		s.l.Error("failed to listen", "err", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}

// ListenTLS starts the Modbus server listening on "address:port".
func (s *Server) ListenTLS(addressPort string, config *tls.Config) (err error) {
	listen, err := tls.Listen("tcp", addressPort, config)
	if err != nil {
		s.l.Error("failed to listen on TLS", "err", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}
