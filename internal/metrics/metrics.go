package metrics

import (
	"math"
	"sort"
	"time"
)

type Stats struct {
	Min time.Duration
	Avg time.Duration
	P50 time.Duration
	P95 time.Duration
	P99 time.Duration
	Max time.Duration
}

type JSONStats struct {
	MinMS float64 `json:"min_ms"`
	AvgMS float64 `json:"avg_ms"`
	P50MS float64 `json:"p50_ms"`
	P95MS float64 `json:"p95_ms"`
	P99MS float64 `json:"p99_ms"`
	MaxMS float64 `json:"max_ms"`
}

func Calculate(values []time.Duration) Stats {
	if len(values) == 0 {
		return Stats{}
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, v := range sorted {
		total += v
	}

	return Stats{
		Min: sorted[0],
		Avg: total / time.Duration(len(sorted)),
		P50: percentile(sorted, 50),
		P95: percentile(sorted, 95),
		P99: percentile(sorted, 99),
		Max: sorted[len(sorted)-1],
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil((p / 100) * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

func (s Stats) JSON() JSONStats {
	return JSONStats{
		MinMS: ms(s.Min),
		AvgMS: ms(s.Avg),
		P50MS: ms(s.P50),
		P95MS: ms(s.P95),
		P99MS: ms(s.P99),
		MaxMS: ms(s.Max),
	}
}

func ms(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}
