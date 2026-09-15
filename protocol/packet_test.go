package protocol

import "testing"

func TestRoundTrip(t *testing.T) {
	for _, want := range []Packet{
		{Type: Ping, Sequence: 42, ClientSendUs: 12345},
		{Type: Pong, Sequence: 42, ClientSendUs: 12345, ServerReceiveUs: 200, ServerSendUs: 300},
	} {
		b, err := Encode(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
}

func TestBadPackets(t *testing.T) {
	good, _ := Encode(Packet{Type: Ping})
	cases := [][]byte{good[:3], good[:len(good)-1]}
	for _, change := range []struct {
		index int
		value byte
	}{{0, 99}, {5, 99}, {4, 9}} {
		b := append([]byte(nil), good...)
		b[change.index] = change.value
		cases = append(cases, b)
	}
	for _, b := range cases {
		if _, err := Decode(b); err == nil {
			t.Fatalf("accepted %v", b)
		}
	}
}
