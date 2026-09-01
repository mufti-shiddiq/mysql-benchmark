package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/benchmark"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/config"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/database"
)

type Report struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Database    DatabaseReport    `json:"database"`
	Benchmark   string            `json:"benchmark"`
	ScaleFactor int               `json:"scale_factor"`
	Config      ConfigReport      `json:"configuration"`
	Results     []BenchmarkResult `json:"results"`
	Summary     Summary           `json:"summary"`
}

type DatabaseReport struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	MySQLVersion string `json:"mysql_version"`
}

type ConfigReport struct {
	Warmup     int `json:"warmup"`
	Iterations int `json:"iterations"`
}

type BenchmarkResult struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Warmup      int     `json:"warmup"`
	Iterations  int     `json:"iterations"`
	MinMS       float64 `json:"min_ms"`
	AvgMS       float64 `json:"avg_ms"`
	P50MS       float64 `json:"p50_ms"`
	P95MS       float64 `json:"p95_ms"`
	P99MS       float64 `json:"p99_ms"`
	MaxMS       float64 `json:"max_ms"`
	Error       string  `json:"error,omitempty"`
}

type Summary struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

func NewReport(cfg config.Config, info database.Info, results []benchmark.Result) Report {
	report := Report{
		GeneratedAt: time.Now().UTC(),
		Database: DatabaseReport{
			Host:         cfg.Host,
			Port:         cfg.Port,
			Database:     cfg.Database,
			MySQLVersion: info.MySQLVersion,
		},
		Benchmark:   cfg.Mode,
		ScaleFactor: cfg.ScaleFactor,
		Config: ConfigReport{
			Warmup:     cfg.Warmup,
			Iterations: cfg.Iterations,
		},
		Results: make([]BenchmarkResult, 0, len(results)),
	}
	for _, r := range results {
		stats := r.Stats.JSON()
		item := BenchmarkResult{
			Name:        r.Name,
			Description: r.Description,
			Warmup:      r.Warmup,
			Iterations:  r.Iterations,
			MinMS:       stats.MinMS,
			AvgMS:       stats.AvgMS,
			P50MS:       stats.P50MS,
			P95MS:       stats.P95MS,
			P99MS:       stats.P99MS,
			MaxMS:       stats.MaxMS,
			Error:       r.Error,
		}
		report.Results = append(report.Results, item)
		report.Summary.Total++
		if r.Error == "" {
			report.Summary.Successful++
		} else {
			report.Summary.Failed++
		}
	}
	return report
}

func WriteText(w io.Writer, report Report) {
	line(w, "Database")
	_, _ = fmt.Fprintf(w, "Host       : %s\n", report.Database.Host)
	_, _ = fmt.Fprintf(w, "Port       : %d\n", report.Database.Port)
	_, _ = fmt.Fprintf(w, "Database   : %s\n", report.Database.Database)
	_, _ = fmt.Fprintf(w, "MySQL      : %s\n", report.Database.MySQLVersion)

	line(w, strings.ToUpper(report.Benchmark))
	_, _ = fmt.Fprintf(w, "%-30s %9s %9s %9s %9s %9s %9s\n", "Test", "Min", "Avg", "P50", "P95", "P99", "Max")
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 91))
	for _, r := range report.Results {
		if r.Error != "" {
			_, _ = fmt.Fprintf(w, "%-30s ERROR: %s\n", r.Name, r.Error)
			continue
		}
		_, _ = fmt.Fprintf(w, "%-30s %8.2fms %8.2fms %8.2fms %8.2fms %8.2fms %8.2fms\n", r.Name, r.MinMS, r.AvgMS, r.P50MS, r.P95MS, r.P99MS, r.MaxMS)
	}

	line(w, "Summary")
	_, _ = fmt.Fprintf(w, "Total tests : %d\nSuccessful  : %d\nFailed      : %d\n", report.Summary.Total, report.Summary.Successful, report.Summary.Failed)
	if report.Summary.Failed == 0 {
		_, _ = fmt.Fprintln(w, "\nBenchmark completed")
	}
}

func WriteJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func WriteOutput(pathOrMode string, textOut io.Writer, report Report) error {
	switch {
	case pathOrMode == "":
		WriteText(textOut, report)
		return nil
	case pathOrMode == "text":
		WriteText(textOut, report)
		return nil
	case pathOrMode == "json":
		return WriteJSON(textOut, report)
	case strings.HasSuffix(strings.ToLower(pathOrMode), ".json"):
		f, err := os.Create(pathOrMode)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := WriteJSON(f, report); err != nil {
			return err
		}
		WriteText(textOut, report)
		_, _ = fmt.Fprintf(textOut, "\nJSON results written to %s\n", pathOrMode)
		return nil
	default:
		return fmt.Errorf("unsupported output %q; use text, json, or a .json file path", pathOrMode)
	}
}

func line(w io.Writer, title string) {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, strings.Repeat("=", 72))
	_, _ = fmt.Fprintln(w, title)
	_, _ = fmt.Fprintln(w, strings.Repeat("=", 72))
	_, _ = fmt.Fprintln(w)
}
