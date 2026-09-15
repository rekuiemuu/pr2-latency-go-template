package telemetry

import (
	"errors"
	"math"
	"sort"
	"time"
)

const Timeout = time.Second
const Capacity = 1024
const (
	Received  = "received"
	TimedOut  = "timeout"
	Late      = "late_response"
	Duplicate = "duplicate_response"
	Unknown   = "unknown_response"
	Invalid   = "invalid_response"
)

type Flight struct {
	Sequence uint16
	Sent     time.Time
	Series   string
	Status   string
	Attempts int
}

type Sample struct {
	Series   string
	Number   int
	Sequence uint16
	SentAtMs float64
	RTTMs    float64
	SRTTMs   float64
	Status   string
}

type Tracker struct {
	flights map[uint16]*Flight
	Samples []Sample
	srtt    float64
}

func New() *Tracker { return &Tracker{flights: make(map[uint16]*Flight)} }

func (t *Tracker) Sent(seq uint16, sent time.Time, series string, number int, start time.Time) error {
	if len(t.flights) >= Capacity {
		return errors.New("inFlight full")
	}
	if _, exists := t.flights[seq]; exists {
		return errors.New("sequence already used")
	}
	t.flights[seq] = &Flight{Sequence: seq, Sent: sent, Series: series, Attempts: 1}
	t.Samples = append(t.Samples, Sample{Series: series, Number: number, Sequence: seq, SentAtMs: sent.Sub(start).Seconds() * 1000})
	return nil
}

func (t *Tracker) Pong(seq uint16, clientSendUs uint64, now time.Time) string {
	f := t.flights[seq]
	if f == nil {
		return Unknown
	}
	if uint64(f.Sent.UnixMicro()) != clientSendUs {
		return Invalid
	}
	if f.Status == Received {
		return Duplicate
	}
	if f.Status == TimedOut || now.Sub(f.Sent) > Timeout {
		if f.Status == "" {
			t.mark(seq, TimedOut, 0)
		}
		return Late
	}
	rtt := now.Sub(f.Sent).Seconds() * 1000
	if rtt < 0 {
		return Invalid
	}
	if t.srtt == 0 {
		t.srtt = rtt
	} else {
		t.srtt = .875*t.srtt + .125*rtt
	}
	t.mark(seq, Received, rtt)
	return Received
}

func (t *Tracker) Expire(now time.Time) {
	for seq, f := range t.flights {
		if f.Status == "" && now.Sub(f.Sent) > Timeout {
			t.mark(seq, TimedOut, 0)
		}
	}
}

func (t *Tracker) mark(seq uint16, status string, rtt float64) {
	t.flights[seq].Status = status
	for i := range t.Samples {
		if t.Samples[i].Sequence == seq {
			t.Samples[i].Status = status
			t.Samples[i].RTTMs = rtt
			if status == Received {
				t.Samples[i].SRTTMs = t.srtt
			}
			break
		}
	}
}

type Stats struct {
	Sent, Received, Timeouts                       int
	Min, Max, Mean, Median, SRTT, Jitter, LossRate float64
}

func Calculate(samples []Sample) Stats {
	s := Stats{Sent: len(samples)}
	var values []float64
	var previous float64
	for _, v := range samples {
		if v.Status == TimedOut {
			s.Timeouts++
		}
		if v.Status != Received {
			continue
		}
		if s.Received > 0 {
			s.Jitter += math.Abs(v.RTTMs - previous)
		}
		previous = v.RTTMs
		s.Received++
		s.Mean += v.RTTMs
		s.SRTT = v.SRTTMs
		values = append(values, v.RTTMs)
	}
	if s.Sent > 0 {
		s.LossRate = float64(s.Timeouts) / float64(s.Sent) * 100
	}
	if s.Received == 0 {
		return s
	}
	sort.Float64s(values)
	s.Min, s.Max = values[0], values[len(values)-1]
	s.Mean /= float64(s.Received)
	s.Median = values[len(values)/2]
	if len(values)%2 == 0 {
		s.Median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	if s.Received > 1 {
		s.Jitter /= float64(s.Received - 1)
	}
	return s
}
