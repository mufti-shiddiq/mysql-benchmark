package benchmark

import (
	"context"
	"time"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/metrics"
)

type Case struct {
	Name        string
	Description string
	Run         func(context.Context) error
}

type Result struct {
	Name        string
	Description string
	Warmup      int
	Iterations  int
	Stats       metrics.Stats
	Error       string
}

type Runner struct {
	Warmup     int
	Iterations int
	Timeout    time.Duration
}

func (r Runner) Run(ctx context.Context, cases []Case) []Result {
	results := make([]Result, 0, len(cases))
	for _, c := range cases {
		result := Result{Name: c.Name, Description: c.Description, Warmup: r.Warmup, Iterations: r.Iterations}
		if c.Run == nil {
			result.Error = "benchmark case has no runner"
			results = append(results, result)
			continue
		}
		if err := r.runWarmup(ctx, c); err != nil {
			result.Error = sanitizeError(err)
			results = append(results, result)
			continue
		}
		durations, err := r.runMeasured(ctx, c)
		if err != nil {
			result.Error = sanitizeError(err)
			results = append(results, result)
			continue
		}
		result.Stats = metrics.Calculate(durations)
		results = append(results, result)
	}
	return results
}

func (r Runner) runWarmup(ctx context.Context, c Case) error {
	for i := 0; i < r.Warmup; i++ {
		caseCtx, cancel := r.caseContext(ctx)
		err := c.Run(caseCtx)
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) runMeasured(ctx context.Context, c Case) ([]time.Duration, error) {
	durations := make([]time.Duration, 0, r.Iterations)
	for i := 0; i < r.Iterations; i++ {
		caseCtx, cancel := r.caseContext(ctx)
		start := time.Now()
		err := c.Run(caseCtx)
		elapsed := time.Since(start)
		cancel()
		if err != nil {
			return durations, err
		}
		durations = append(durations, elapsed)
	}
	return durations, nil
}

func (r Runner) caseContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
