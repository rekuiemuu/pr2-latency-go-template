package protocol

import "errors"

const Version uint16 = 1
const HeaderSize = 7
const (
	Ping     byte = 1
	Pong     byte = 2
	Movement byte = 3
	Shoot    byte = 4
)

type Packet struct {
	Type            byte
	Sequence        uint16
	ClientSendUs    uint64
	ServerReceiveUs uint64
	ServerSendUs    uint64
}

// TODO: записать поля в network byte order, без побайтового копирования struct.
func Encode(p Packet) ([]byte, error) { return nil, errors.New("TODO: Encode") }

// TODO: проверить заголовок, тип, версию, размер и длину payload до чтения.
func Decode(b []byte) (Packet, error) { return Packet{}, errors.New("TODO: Decode") }
