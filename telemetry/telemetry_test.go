package telemetry

import (
	"testing"
	"time"
)

func TestSRTTAndTimeout(t *testing.T) {
	start := time.Unix(100, 0)
	x := New()
	if err := x.Sent(1, start, "baseline", 1, start); err != nil {
		t.Fatal(err)
	}
	if got := x.Pong(1, uint64(start.UnixMicro()), start.Add(100*time.Millisecond)); got != Received {
		t.Fatal(got)
	}
	if x.Samples[0].RTTMs != 100 || x.Samples[0].SRTTMs != 100 {
		t.Fatal(x.Samples)
	}
}

func TestStats(t *testing.T) {
	s := Calculate([]Sample{{RTTMs: 10, Status: Received}, {RTTMs: 20, Status: Received}, {Status: TimedOut}})
	if s.Sent != 3 || s.Received != 2 || s.Mean != 15 || s.Jitter != 10 {
		t.Fatal(s)
	}
}
