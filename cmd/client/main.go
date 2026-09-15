package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"go-algo/protocol"
	"go-algo/telemetry"
	"go-algo/transport"
	"log"
	"os"
	"strconv"
	"time"
)

type response struct {
	data []byte
	at   time.Time
}

func main() {
	series := flag.String("series", "baseline", "experiment id")
	address := flag.String("server", "127.0.0.1:19000", "UDP server address")
	count := flag.Int("count", 50, "number of PINGs")
	interval := flag.Duration("interval", 200*time.Millisecond, "PING interval")
	output := flag.String("csv", "docs/latency_samples.csv", "CSV output")
	flag.Parse()
	if *count < 1 || *count > 1000 || *interval < 200*time.Millisecond || *interval > time.Second {
		log.Fatal("count must be 1..1000; interval must be 200ms..1s")
	}
	socket, err := transport.Dial(*address)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	tracker := telemetry.New()
	answers := make(chan response, 100)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			b, err := socket.Receive(100 * time.Millisecond)
			if err == nil {
				answers <- response{b, time.Now()}
			}
		}
	}()
	start := time.Now()
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	sent := 0
	for sent < *count {
		select {
		case <-ticker.C:
			sent++
			now := time.Now()
			seq := uint16(sent)
			if err := tracker.Sent(seq, now, *series, sent, start); err != nil {
				log.Fatal(err)
			}
			b, _ := protocol.Encode(protocol.Packet{Type: protocol.Ping, Sequence: seq, ClientSendUs: uint64(now.UnixMicro())})
			if err := socket.Send(b); err != nil {
				log.Fatal(err)
			}
		case answer := <-answers:
			handle(tracker, answer)
		}
		tracker.Expire(time.Now())
	}
	until := time.NewTimer(telemetry.Timeout + 200*time.Millisecond)
	for {
		select {
		case answer := <-answers:
			handle(tracker, answer)
		case <-until.C:
			close(done)
			tracker.Expire(time.Now())
			save(*output, tracker.Samples)
			s := telemetry.Calculate(tracker.Samples)
			fmt.Printf("%s: sent=%d received=%d timeout=%d mean=%.3fms srtt=%.3fms jitter=%.3fms loss=%.1f%%\n", *series, s.Sent, s.Received, s.Timeouts, s.Mean, s.SRTT, s.Jitter, s.LossRate)
			return
		}
	}
}

func handle(t *telemetry.Tracker, a response) {
	p, err := protocol.Decode(a.data)
	if err != nil {
		log.Printf("invalid_response: %v", err)
		return
	}
	if p.Type != protocol.Pong {
		log.Printf("invalid_response: not PONG")
		return
	}
	status := t.Pong(p.Sequence, p.ClientSendUs, a.at)
	if status != telemetry.Received {
		log.Printf("seq=%d %s", p.Sequence, status)
	}
}

func save(path string, samples []telemetry.Sample) {
	if err := os.MkdirAll("docs", 0755); err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	info, _ := f.Stat()
	w := csv.NewWriter(f)
	w.Comma = ';'
	if info.Size() == 0 {
		w.Write([]string{"experiment_id", "sample", "sequence", "sent_at_ms", "rtt_ms", "srtt_ms", "status"})
	}
	for _, s := range samples {
		rtt, srtt := "", ""
		if s.Status == telemetry.Received {
			rtt = fmt.Sprintf("%.3f", s.RTTMs)
			srtt = fmt.Sprintf("%.3f", s.SRTTMs)
		}
		w.Write([]string{s.Series, strconv.Itoa(s.Number), strconv.Itoa(int(s.Sequence)), fmt.Sprintf("%.3f", s.SentAtMs), rtt, srtt, s.Status})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}
