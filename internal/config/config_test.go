package config

import "testing"

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
