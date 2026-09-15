package telemetry

import (
	"errors"
	"time"
)

const Timeout = time.Second
const (
	Received  = "received"
	TimedOut  = "timeout"
	Late      = "late_response"
	Duplicate = "duplicate_response"
	Unknown   = "unknown_response"
	Invalid   = "invalid_response"
)

type Sample struct {
	Series                  string
	Number                  int
	Sequence                uint16
	SentAtMs, RTTMs, SRTTMs float64
	Status                  string
}

type Tracker struct{ Samples []Sample }

func New() *Tracker { return &Tracker{} }

// TODO: ограниченный inFlight: номер, время, серия, статус и число попыток.
func (t *Tracker) Sent(seq uint16, sent time.Time, series string, number int, start time.Time) error {
	return errors.New("TODO: Sent")
}

// TODO: RTT только из монотонной разности клиентских time.Time; SRTT с alpha=0.125.
func (t *Tracker) Pong(seq uint16, clientSendUs uint64, now time.Time) string { return Invalid }

// TODO: отметить ответы старше 1000 мс как timeout.
func (t *Tracker) Expire(now time.Time) {}

type Stats struct {
	Sent, Received, Timeouts                       int
	Min, Max, Mean, Median, SRTT, Jitter, LossRate float64
}

// TODO: min/max/mean/median, средний |RTT[i]-RTT[i-1]| и timeout/sent*100.
func Calculate(samples []Sample) Stats { return Stats{} }
