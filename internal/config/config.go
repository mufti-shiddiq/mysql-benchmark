package config

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
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
	EnvFile     string
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
	return ParseWithDotEnv(args, getenv, nil)
}

func ParseWithDotEnv(args []string, getenv func(string) string, readFile func(string) ([]byte, error)) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	cfg := Defaults()
	cfg.EnvFile = envFileArg(args)
	if cfg.EnvFile == "" {
		cfg.EnvFile = ".env"
	}
	if readFile != nil {
		values, err := loadDotEnv(cfg.EnvFile, readFile)
		if err != nil {
			return cfg, err
		}
		applyValues(&cfg, values)
	}
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
	fs.StringVar(&cfg.EnvFile, "env-file", cfg.EnvFile, "Path to .env file")
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
	values := map[string]string{}
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD"} {
		if v := getenv(key); v != "" {
			values[key] = v
		}
	}
	applyValues(cfg, values)
}

func applyValues(cfg *Config, values map[string]string) {
	if v := values["DB_HOST"]; v != "" {
		cfg.Host = v
	}
	if v := values["DB_PORT"]; v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := values["DB_NAME"]; v != "" {
		cfg.Database = v
	}
	if v := values["DB_USER"]; v != "" {
		cfg.User = v
	}
	if v := values["DB_PASSWORD"]; v != "" {
		cfg.Password = v
	}
}

func envFileArg(args []string) string {
	for i, arg := range args {
		if arg == "--env-file" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--env-file=") {
			return strings.TrimPrefix(arg, "--env-file=")
		}
	}
	return ".env"
}

func loadDotEnv(path string, readFile func(string) ([]byte, error)) (map[string]string, error) {
	data, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("unable to read env file %q: %w", path, err)
	}
	values := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		values[key] = unquote(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}
