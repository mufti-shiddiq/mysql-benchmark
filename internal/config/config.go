package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const Version = "0.1.0"

type Config struct {
	Host        string
	Port        int
	Database    string
	User        string
	Password    string
	Mode        string
	ScaleFactor int
	Warmup      int
	Iterations  int
	Output      string
	Force       bool
	ShowVersion bool
}

func Defaults() Config {
	return Config{
		Port:        3306,
		Mode:        "sakila",
		ScaleFactor: 1,
		Warmup:      5,
		Iterations:  30,
	}
}

func Parse(args []string, getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	cfg := Defaults()
	applyEnv(&cfg, getenv)

	fs := flag.NewFlagSet("mysql-benchmark", flag.ContinueOnError)
	fs.StringVar(&cfg.Host, "host", cfg.Host, "MySQL host")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "MySQL port")
	fs.StringVar(&cfg.Database, "database", cfg.Database, "MySQL database name")
	fs.StringVar(&cfg.User, "user", cfg.User, "MySQL username")
	fs.StringVar(&cfg.Mode, "mode", cfg.Mode, "Benchmark mode: sakila, tpch, both")
	fs.IntVar(&cfg.ScaleFactor, "scale-factor", cfg.ScaleFactor, "TPC-H-inspired scale factor")
	fs.IntVar(&cfg.Warmup, "warmup", cfg.Warmup, "Warmup iterations per benchmark case")
	fs.IntVar(&cfg.Iterations, "iterations", cfg.Iterations, "Measured iterations per benchmark case")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "Output mode: text, json, or a .json file path")
	fs.BoolVar(&cfg.Force, "force", cfg.Force, "Accept destructive benchmark setup without prompting")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Print version and exit")

	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	cfg.Output = strings.TrimSpace(cfg.Output)
	return cfg, cfg.ValidatePartial()
}

func (c Config) ValidatePartial() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d", c.Port)
	}
	if c.Warmup < 0 {
		return errors.New("warmup must be zero or greater")
	}
	if c.Iterations <= 0 {
		return errors.New("iterations must be greater than zero")
	}
	if c.ScaleFactor <= 0 {
		return errors.New("scale-factor must be greater than zero")
	}
	switch c.Mode {
	case "", "sakila", "tpch", "both":
	default:
		return fmt.Errorf("invalid mode %q", c.Mode)
	}
	return nil
}

func (c Config) ValidateRequired() error {
	var missing []string
	if c.Host == "" {
		missing = append(missing, "host")
	}
	if c.Database == "" {
		missing = append(missing, "database")
	}
	if c.User == "" {
		missing = append(missing, "user")
	}
	if c.Password == "" {
		missing = append(missing, "password")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}
	return c.ValidatePartial()
}

func applyEnv(cfg *Config, getenv func(string) string) {
	if v := getenv("DB_HOST"); v != "" {
		cfg.Host = v
	}
	if v := getenv("DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := getenv("DB_NAME"); v != "" {
		cfg.Database = v
	}
	if v := getenv("DB_USER"); v != "" {
		cfg.User = v
	}
	if v := getenv("DB_PASSWORD"); v != "" {
		cfg.Password = v
	}
}
