package telemetry

import (
	"math"
	"testing"
	"time"
)

func TestStatesAndSRTT(t *testing.T) {
	start := time.Unix(100, 0)
	x := New()
	x.Sent(1, start, "baseline", 1, start)
	x.Sent(2, start, "baseline", 2, start)
	if got := x.Pong(1, uint64(start.UnixMicro()), start.Add(100*time.Millisecond)); got != Received {
		t.Fatal(got)
	}
	if got := x.Pong(1, uint64(start.UnixMicro()), start.Add(110*time.Millisecond)); got != Duplicate {
		t.Fatal(got)
	}
	if got := x.Pong(99, 0, start); got != Unknown {
		t.Fatal(got)
	}
	if got := x.Pong(2, 0, start); got != Invalid {
		t.Fatal(got)
	}
	if got := x.Pong(2, uint64(start.UnixMicro()), start.Add(1200*time.Millisecond)); got != Late {
		t.Fatal(got)
	}
	if x.Samples[1].Status != TimedOut {
		t.Fatal(x.Samples)
	}
	x.Sent(3, start, "baseline", 3, start)
	x.Pong(3, uint64(start.UnixMicro()), start.Add(200*time.Millisecond))
	if math.Abs(x.Samples[2].SRTTMs-112.5) > 0.001 {
		t.Fatal(x.Samples[2])
	}
}

func TestStats(t *testing.T) {
	x := []Sample{{RTTMs: 10, SRTTMs: 10, Status: Received}, {RTTMs: 20, SRTTMs: 11.25, Status: Received}, {Status: TimedOut}}
	s := Calculate(x)
	if s.Sent != 3 || s.Received != 2 || s.Timeouts != 1 || s.Mean != 15 || s.Median != 15 || s.Jitter != 10 || math.Abs(s.LossRate-100.0/3) > 0.001 {
		t.Fatal(s)
	}
}
