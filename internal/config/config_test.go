package config

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Port != 3306 {
		t.Fatalf("Port = %d, want 3306", cfg.Port)
	}
	if cfg.Warmup != 5 || cfg.Iterations != 30 {
		t.Fatalf("unexpected iteration defaults: warmup=%d iterations=%d", cfg.Warmup, cfg.Iterations)
	}
}

func TestParseEnvironment(t *testing.T) {
	env := map[string]string{
		"DB_HOST":     "db.example.com",
		"DB_PORT":     "3307",
		"DB_NAME":     "benchmark",
		"DB_USER":     "bench",
		"DB_PASSWORD": "secret",
	}
	cfg, err := Parse(nil, func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "db.example.com" || cfg.Port != 3307 || cfg.Database != "benchmark" || cfg.User != "bench" || cfg.Password != "secret" {
		t.Fatalf("environment was not applied: %+v", cfg)
	}
}

func TestParseFlagsOverrideEnvironment(t *testing.T) {
	env := map[string]string{"DB_HOST": "env-host", "DB_PORT": "3306", "DB_NAME": "envdb", "DB_USER": "envuser"}
	cfg, err := Parse([]string{"--host", "flag-host", "--port", "4406", "--database", "flagdb", "--user", "flaguser", "--mode", "both"}, func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "flag-host" || cfg.Port != 4406 || cfg.Database != "flagdb" || cfg.User != "flaguser" || cfg.Mode != "both" {
		t.Fatalf("flags did not override env: %+v", cfg)
	}
}

func TestValidateRequired(t *testing.T) {
	cfg := Defaults()
	if err := cfg.ValidateRequired(); err == nil {
		t.Fatal("ValidateRequired() error = nil, want missing values error")
	}
}

func TestParseDotEnv(t *testing.T) {
	files := map[string]string{
		".env": strings.Join([]string{
			"DB_HOST=dotenv-host",
			"DB_PORT=3308",
			"DB_NAME=dotenvdb",
			"DB_USER=dotenvuser",
			"DB_PASSWORD='dotenv secret'",
		}, "\n"),
	}
	cfg, err := ParseWithDotEnv(nil, func(string) string { return "" }, func(path string) ([]byte, error) {
		value, ok := files[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return []byte(value), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "dotenv-host" || cfg.Port != 3308 || cfg.Database != "dotenvdb" || cfg.User != "dotenvuser" || cfg.Password != "dotenv secret" {
		t.Fatalf(".env was not applied: %+v", cfg)
	}
}

func TestParseDotEnvPrecedence(t *testing.T) {
	env := map[string]string{"DB_HOST": "env-host", "DB_PORT": "3309"}
	cfg, err := ParseWithDotEnv([]string{"--host", "flag-host", "--env-file", "custom.env"}, func(k string) string {
		return env[k]
	}, func(path string) ([]byte, error) {
		if path != "custom.env" {
			t.Fatalf("read env path = %q, want custom.env", path)
		}
		return []byte("DB_HOST=dotenv-host\nDB_PORT=3308\nDB_NAME=dotenvdb\nDB_USER=dotenvuser\nDB_PASSWORD=dotenvpass\n"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "flag-host" {
		t.Fatalf("Host = %q, want flag-host", cfg.Host)
	}
	if cfg.Port != 3309 {
		t.Fatalf("Port = %d, want shell env override", cfg.Port)
	}
	if cfg.Database != "dotenvdb" || cfg.User != "dotenvuser" || cfg.Password != "dotenvpass" {
		t.Fatalf(".env fallback values not applied: %+v", cfg)
	}
}

func TestParseDotEnvReadError(t *testing.T) {
	_, err := ParseWithDotEnv(nil, func(string) string { return "" }, func(string) ([]byte, error) {
		return nil, errors.New("permission denied")
	})
	if err == nil {
		t.Fatal("ParseWithDotEnv() error = nil, want read error")
	}
}
