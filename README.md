# MySQL Benchmark

A local CLI for comparing MySQL database performance from the machine where the benchmark runs, typically an Ubuntu VPS near your application.

## Overview

`mysql-benchmark` measures connection/query latency, relational query performance, joins, aggregation, writes, transactions, and analytical workloads. It is designed for comparative testing from the same client location to different MySQL providers.

## Features

- Interactive CLI and non-interactive flags.
- Environment variable support for automation.
- Safe database validation before setup.
- Explicit confirmation for non-empty databases unless `--force` is used.
- Warmup and repeated measured iterations.
- Min, average, p50, p95, p99, and max metrics.
- Sakila-style relational workload.
- TPC-H-inspired analytical workload.
- Terminal table output and JSON output.
- No telemetry and no automatic result upload.

## Quick Start

```bash
git clone https://github.com/mufti-shiddiq/mysql-benchmark.git
cd mysql-benchmark
go build -o mysql-benchmark ./cmd/mysql-benchmark
./mysql-benchmark
```

## Requirements

- Go 1.22 or newer for building from source.
- A reachable MySQL-compatible database.
- An existing database/schema name. The CLI does not create databases automatically.
- A user with permission to create, alter, insert into, update, delete from, and drop benchmark-owned tables.

## Interactive Usage

```bash
./mysql-benchmark
```

The CLI prompts for host, port, database, username, password, benchmark mode, warmup, and iterations when values are missing.

## CLI Usage

```bash
DB_PASSWORD='secret' ./mysql-benchmark \
  --host db.example.com \
  --port 3306 \
  --database benchmark \
  --user benchmark \
  --mode both \
  --warmup 5 \
  --iterations 30
```

Flags:

```text
--host
--port
--database
--user
--mode sakila|tpch|both
--scale-factor 1
--warmup
--iterations
--output json|results.json
--force
--help
--version
```

Do not pass passwords in shell arguments. Use the interactive password prompt or `DB_PASSWORD`.

## Benchmark Modes

### Sakila

Creates a Sakila-style relational dataset with benchmark-owned tables. It tests round trips, simple indexed selects, two-table joins, multi-table joins, complex joins, aggregation, sorting, subqueries, inserts, updates, deletes, and transactions.

### TPC-H

Creates a TPC-H-inspired analytical dataset with benchmark-owned tables and runs query cases named Q01 through Q22 where practical for MySQL. This is not an official TPC-H result and should not be described as TPC-H compliant.

## Benchmark Methodology

Each benchmark case runs warmup iterations first. Warmup timings are discarded. The measured iterations are then timed and summarized with min, average, p50, p95, p99, and max in milliseconds.

Setup and dataset loading time are reported separately from query timings.

`SELECT 1 latency` represents client/server round-trip time plus MySQL processing overhead. It is not pure TCP network RTT.

## Understanding Results

Results depend on network distance, MySQL configuration, indexes, storage engine, CPU, RAM, dataset size, and cache state. Treat numbers as comparative signals, not absolute production guarantees.

## Comparing Database Providers

Run comparisons from the same machine and location:

```text
Ubuntu VPS
  -> Provider A MySQL
  -> Provider B MySQL
```

Do not compare `Laptop -> Database A` with `VPS -> Database B` for network-sensitive measurements.

## Security

- Passwords are never printed.
- JSON output does not include passwords.
- DSNs containing credentials are not logged.
- Non-empty databases require confirmation unless `--force` is set.
- Benchmark tables use explicit prefixes.

## Third-party Datasets

This repository does not redistribute Sakila SQL files, TPC-H tools, or generated TPC-H data. See [THIRD_PARTY_LICENSES.md](./THIRD_PARTY_LICENSES.md).

## Development

```bash
make build
make test
make lint
go build ./...
```

Cross-compile examples:

```bash
GOOS=linux GOARCH=amd64 go build -o dist/mysql-benchmark-linux-amd64 ./cmd/mysql-benchmark
GOOS=linux GOARCH=arm64 go build -o dist/mysql-benchmark-linux-arm64 ./cmd/mysql-benchmark
```

## Roadmap

v0.2:

- concurrency benchmarks
- configurable concurrency
- CSV output
- larger TPC-H scale factors
- benchmark comparison mode

v0.3:

- historical benchmark results
- regression detection
- richer reports
- HTML report

Future:

- hosted result dashboard
- benchmark result sharing

## License

MIT. See [LICENSE](./LICENSE).
