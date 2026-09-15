package main

import (
	"flag"
	"go-algo/protocol"
	"go-algo/transport"
	"log"
	"math/rand"
	"time"
)

func main() {
	mode := flag.String("mode", "baseline", "baseline, delay_50, delay_100, jitter, loss_5, combined")
	address := flag.String("addr", "127.0.0.1:19000", "UDP listen address")
	seed := flag.Int64("seed", 42, "jitter seed")
	flag.Parse()
	server, err := transport.Listen(*address)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	rng := rand.New(rand.NewSource(*seed))
	log.Printf("listening %s mode=%s seed=%d", *address, *mode, *seed)
	for {
		b, client, err := server.Receive()
		if err != nil {
			log.Print(err)
			continue
		}
		ping, err := protocol.Decode(b)
		if err != nil {
			log.Printf("invalid datagram from %s: %v", client, err)
			continue
		}
		if ping.Type != protocol.Ping {
			log.Printf("unexpected PONG from %s", client)
			continue
		}
		received := uint64(time.Now().UnixMicro())
		log.Printf("PING seq=%d from %s", ping.Sequence, client)
		if (*mode == "loss_5" || *mode == "combined") && ping.Sequence%20 == 0 {
			continue
		}
		delay := time.Duration(0)
		switch *mode {
		case "delay_50":
			delay = 50 * time.Millisecond
		case "delay_100":
			delay = 100 * time.Millisecond
		case "jitter":
			delay = time.Duration(50+rng.Intn(101)) * time.Millisecond
		case "combined":
			delay = time.Duration(100+rng.Intn(101)) * time.Millisecond
		case "baseline", "loss_5":
		default:
			log.Printf("unknown mode %q", *mode)
			continue
		}
		go func() {
			time.Sleep(delay)
			pong, _ := protocol.Encode(protocol.Packet{Type: protocol.Pong, Sequence: ping.Sequence, ClientSendUs: ping.ClientSendUs, ServerReceiveUs: received, ServerSendUs: uint64(time.Now().UnixMicro())})
			if err := server.Send(pong, client); err != nil {
				log.Print(err)
			}
		}()
	}
}
