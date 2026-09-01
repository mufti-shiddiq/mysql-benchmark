package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/benchmark"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/config"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/database"
)

func TestJSONDoesNotIncludePassword(t *testing.T) {
	cfg := config.Config{
		Host:       "db.example.com",
		Port:       3306,
		Database:   "benchmark",
		User:       "bench",
		Password:   "super-secret",
		Mode:       "sakila",
		Warmup:     5,
		Iterations: 30,
	}
	report := NewReport(cfg, database.Info{MySQLVersion: "8.0.test"}, []benchmark.Result{{Name: "select_1", Warmup: 5, Iterations: 30}})
	var buf bytes.Buffer
	if err := WriteJSON(&buf, report); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "super-secret") || strings.Contains(out, "DB_PASSWORD") {
		t.Fatalf("JSON leaked secret: %s", out)
	}
	if !strings.Contains(out, `"mysql_version"`) || !strings.Contains(out, `"results"`) {
		t.Fatalf("JSON missing expected fields: %s", out)
	}
}
