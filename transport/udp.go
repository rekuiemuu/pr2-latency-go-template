package transport

import (
	"errors"
	"net"
	"time"
)

type UDP struct{ Conn net.Conn }
type Server struct{ Conn net.PacketConn }

// TODO: использовать системные UDP-сокеты через net.Dial/ListenPacket.
func Dial(address string) (*UDP, error)                      { return nil, errors.New("TODO: Dial") }
func Listen(address string) (*Server, error)                 { return nil, errors.New("TODO: Listen") }
func (u *UDP) Close() error                                  { return nil }
func (u *UDP) Send(b []byte) error                           { return errors.New("TODO: Send") }
func (u *UDP) Receive(timeout time.Duration) ([]byte, error) { return nil, errors.New("TODO: Receive") }
func (s *Server) Close() error                               { return nil }
func (s *Server) Receive() ([]byte, net.Addr, error)         { return nil, nil, errors.New("TODO: Receive") }
func (s *Server) Send(b []byte, addr net.Addr) error         { return errors.New("TODO: Send") }
