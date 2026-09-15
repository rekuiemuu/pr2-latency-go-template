package protocol

import (
	"encoding/binary"
	"errors"
)

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

func Encode(p Packet) ([]byte, error) {
	size := 0
	switch p.Type {
	case Ping:
		size = 8
	case Pong:
		size = 24
	default:
		return nil, errors.New("unsupported packet type")
	}
	b := make([]byte, HeaderSize+size)
	b[0] = p.Type
	binary.BigEndian.PutUint16(b[1:3], p.Sequence)
	binary.BigEndian.PutUint16(b[3:5], uint16(size))
	binary.BigEndian.PutUint16(b[5:7], Version)
	binary.BigEndian.PutUint64(b[7:15], p.ClientSendUs)
	if p.Type == Pong {
		binary.BigEndian.PutUint64(b[15:23], p.ServerReceiveUs)
		binary.BigEndian.PutUint64(b[23:31], p.ServerSendUs)
	}
	return b, nil
}

func Decode(b []byte) (Packet, error) {
	if len(b) < HeaderSize {
		return Packet{}, errors.New("short header")
	}
	kind := b[0]
	if kind != Ping && kind != Pong {
		return Packet{}, errors.New("invalid packet type")
	}
	if binary.BigEndian.Uint16(b[5:7]) != Version {
		return Packet{}, errors.New("invalid version")
	}
	size := int(binary.BigEndian.Uint16(b[3:5]))
	if size != len(b)-HeaderSize {
		return Packet{}, errors.New("payload size mismatch")
	}
	if kind == Ping && size != 8 || kind == Pong && size != 24 {
		return Packet{}, errors.New("invalid payload length")
	}
	p := Packet{Type: kind, Sequence: binary.BigEndian.Uint16(b[1:3]), ClientSendUs: binary.BigEndian.Uint64(b[7:15])}
	if kind == Pong {
		p.ServerReceiveUs = binary.BigEndian.Uint64(b[15:23])
		p.ServerSendUs = binary.BigEndian.Uint64(b[23:31])
	}
	return p, nil
}
