package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/benchmark"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/config"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/database"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/datasets"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/output"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/safety"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR\n\n%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Parse(os.Args[1:], os.Getenv)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if cfg.ShowVersion {
		fmt.Println(config.Version)
		return nil
	}

	statusOut := io.Writer(os.Stdout)
	if cfg.Output == "json" {
		statusOut = os.Stderr
	}

	printHeader(statusOut)
	reader := bufio.NewReader(os.Stdin)
	if err := promptMissing(&cfg, reader); err != nil {
		return err
	}
	if err := cfg.ValidateRequired(); err != nil {
		return err
	}

	ctx := context.Background()
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("unable to configure MySQL connection: %w", err)
	}
	defer db.Close()

	fmt.Fprintln(statusOut, "\nDatabase validation")
	info, err := database.Validate(ctx, db, cfg)
	if err != nil {
		return err
	}
	fmt.Fprintln(statusOut, "Connected to MySQL")
	fmt.Fprintf(statusOut, "MySQL version: %s\n", info.MySQLVersion)
	fmt.Fprintf(statusOut, "Database: %s\n", info.Database)
	if info.TableCount == 0 {
		fmt.Fprintln(statusOut, "Database is empty")
	} else {
		fmt.Fprintf(statusOut, "Database contains %d table(s)\n", info.TableCount)
	}

	ok, err := safety.ConfirmNonEmpty(os.Stdin, statusOut, cfg.Database, info.TableCount, cfg.Force)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("benchmark cancelled")
	}

	allResults, err := runSelectedBenchmarks(ctx, db, cfg, statusOut)
	if err != nil {
		return err
	}
	report := output.NewReport(cfg, info, allResults)
	return output.WriteOutput(cfg.Output, os.Stdout, report)
}

func printHeader(w io.Writer) {
	fmt.Fprintln(w, "================================================")
	fmt.Fprintln(w, "MySQL Benchmark")
	fmt.Fprintln(w, "Database performance testing tool")
	fmt.Fprintln(w, "================================================")
}

func promptMissing(cfg *config.Config, reader *bufio.Reader) error {
	var err error
	if cfg.Host == "" {
		cfg.Host, err = promptString(reader, "Host")
		if err != nil {
			return err
		}
	}
	if cfg.Port == 0 {
		cfg.Port = 3306
	}
	if cfg.Database == "" {
		cfg.Database, err = promptString(reader, "Database")
		if err != nil {
			return err
		}
	}
	if cfg.User == "" {
		cfg.User, err = promptString(reader, "Username")
		if err != nil {
			return err
		}
	}
	if cfg.Password == "" {
		cfg.Password, err = promptPassword(reader, "Password")
		if err != nil {
			return err
		}
	}
	if cfg.Mode == "" {
		cfg.Mode = "sakila"
	}
	if cfg.Mode == "sakila" && len(os.Args) == 1 {
		mode, err := promptMode(reader)
		if err != nil {
			return err
		}
		cfg.Mode = mode
	}
	return nil
}

func promptString(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptPassword(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptMode(reader *bufio.Reader) (string, error) {
	fmt.Println("\nSelect benchmark:")
	fmt.Println("  1. Sakila")
	fmt.Println("  2. TPC-H-inspired")
	fmt.Println("  3. Both")
	fmt.Print("Select [1-3]: ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	switch strings.TrimSpace(value) {
	case "1", "":
		return "sakila", nil
	case "2":
		return "tpch", nil
	case "3":
		return "both", nil
	default:
		return "", fmt.Errorf("invalid benchmark selection")
	}
}

func runSelectedBenchmarks(ctx context.Context, db *sql.DB, cfg config.Config, statusOut io.Writer) ([]benchmark.Result, error) {
	runner := benchmark.Runner{
		Warmup:     cfg.Warmup,
		Iterations: cfg.Iterations,
		Timeout:    30 * time.Second,
	}
	var results []benchmark.Result

	if cfg.Mode == "sakila" || cfg.Mode == "both" {
		fmt.Fprintln(statusOut, "\nPreparing Sakila dataset...")
		if err := datasets.PrepareSakila(ctx, db, cfg.Force); err != nil {
			return nil, fmt.Errorf("prepare Sakila dataset: %w", err)
		}
		fmt.Fprintln(statusOut, "Sakila dataset ready")
		fmt.Fprintln(statusOut, "Running Sakila benchmark...")
		results = append(results, runner.Run(ctx, datasets.SakilaCases(db))...)
	}

	if cfg.Mode == "tpch" || cfg.Mode == "both" {
		fmt.Fprintln(statusOut, "\nPreparing TPC-H-inspired dataset...")
		if err := datasets.PrepareTPCH(ctx, db, cfg.Force, cfg.ScaleFactor); err != nil {
			return nil, fmt.Errorf("prepare TPC-H-inspired dataset: %w", err)
		}
		fmt.Fprintln(statusOut, "TPC-H-inspired dataset ready")
		fmt.Fprintln(statusOut, "Running TPC-H-inspired benchmark...")
		results = append(results, runner.Run(ctx, datasets.TPCHCases(db))...)
	}

	return results, nil
}
