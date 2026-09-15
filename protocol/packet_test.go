package protocol

import "testing"

func TestPingRoundTrip(t *testing.T) {
	want := Packet{Type: Ping, Sequence: 42, ClientSendUs: 12345}
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

func TestShortPacketRejected(t *testing.T) {
	if _, err := Decode([]byte{Ping}); err == nil {
		t.Fatal("short packet accepted")
	}
}
