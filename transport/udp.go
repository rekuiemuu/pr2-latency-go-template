package transport

import (
	"net"
	"time"
)

// UDP contains only socket I/O; packet parsing and metrics live elsewhere.
type UDP struct{ Conn net.Conn }

type Server struct{ Conn net.PacketConn }

func Listen(address string) (*Server, error) {
	c, err := net.ListenPacket("udp", address)
	if err != nil {
		return nil, err
	}
	return &Server{Conn: c}, nil
}
func (s *Server) Close() error { return s.Conn.Close() }
func (s *Server) Receive() ([]byte, net.Addr, error) {
	b := make([]byte, 1024)
	n, addr, err := s.Conn.ReadFrom(b)
	return b[:n], addr, err
}
func (s *Server) Send(b []byte, addr net.Addr) error { _, err := s.Conn.WriteTo(b, addr); return err }

func Dial(address string) (*UDP, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}
	return &UDP{Conn: c}, nil
}

func (u *UDP) Close() error        { return u.Conn.Close() }
func (u *UDP) Send(b []byte) error { _, err := u.Conn.Write(b); return err }
func (u *UDP) Receive(timeout time.Duration) ([]byte, error) {
	u.Conn.SetReadDeadline(time.Now().Add(timeout))
	b := make([]byte, 1024)
	n, err := u.Conn.Read(b)
	return b[:n], err
}
