package metrics

import (
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	values := []time.Duration{
		10 * time.Millisecond,
		1 * time.Millisecond,
		5 * time.Millisecond,
		2 * time.Millisecond,
		100 * time.Millisecond,
	}
	got := Calculate(values)
	if got.Min != 1*time.Millisecond {
		t.Fatalf("Min = %v", got.Min)
	}
	if got.Max != 100*time.Millisecond {
		t.Fatalf("Max = %v", got.Max)
	}
	if got.Avg != 23600*time.Microsecond {
		t.Fatalf("Avg = %v", got.Avg)
	}
	if got.P50 != 5*time.Millisecond {
		t.Fatalf("P50 = %v", got.P50)
	}
	if got.P95 != 100*time.Millisecond || got.P99 != 100*time.Millisecond {
		t.Fatalf("high percentiles = %v %v", got.P95, got.P99)
	}
}

func TestEmptyCalculate(t *testing.T) {
	if got := Calculate(nil); got != (Stats{}) {
		t.Fatalf("Calculate(nil) = %+v", got)
	}
}
