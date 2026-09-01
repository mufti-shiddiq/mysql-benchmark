# MySQL Benchmark

A local CLI for comparing MySQL database performance from the machine where the benchmark runs, typically an Ubuntu VPS near your application.

## Overview

`mysql-benchmark` measures connection/query latency, relational query performance, joins, aggregation, writes, transactions, and analytical workloads. It is designed for comparative testing from the same client location to different MySQL providers.

## Features

- Interactive CLI and non-interactive flags.
- `.env`, environment variable, and CLI flag support for automation.
- Safe database validation before setup.
- Explicit confirmation for non-empty databases unless `--force` is used.
- Warmup and repeated measured iterations.
- Min, average, p50, p95, p99, and max metrics.
- Sakila-style relational workload.
- TPC-H-inspired analytical workload.
- Terminal table output and JSON output.
- No telemetry and no automatic result upload.

## Quick Start

Install the latest release binary:

```bash
curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh | sh
mysql-benchmark
```

Or install to a user-writable directory:

```bash
INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh)"
```

Build from source when developing:

```bash
git clone https://github.com/mufti-shiddiq/mysql-benchmark.git
cd mysql-benchmark
go build -o mysql-benchmark ./cmd/mysql-benchmark
./mysql-benchmark
```

## Requirements

- A reachable MySQL-compatible database.
- An existing database/schema name. The CLI does not create databases automatically.
- A user with permission to create, alter, insert into, update, delete from, and drop benchmark-owned tables.
- Go 1.22 or newer only when building from source.

## Interactive Usage

```bash
./mysql-benchmark
```

The CLI prompts for host, port, database, username, password, benchmark mode, warmup, and iterations when values are missing.

## .env Usage

Copy `.env.example` to `.env` and adjust the connection values:

```bash
cp .env.example .env
```

The CLI reads `.env` by default when it exists. You can also pass a custom file:

```bash
./mysql-benchmark --env-file production.env --mode sakila
```

Configuration precedence is:

```text
defaults < .env file < shell environment < CLI flags
```

Avoid storing passwords in `.env` on shared machines. Prefer the interactive password prompt for local use, or `DB_PASSWORD` for automation when needed.

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
--env-file .env
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

## Troubleshooting

For a local MySQL server, prefer TCP explicitly:

```bash
./mysql-benchmark --host 127.0.0.1 --port 3306 --database benchmark --user root
```

If you see access denied, verify the password and confirm that the MySQL user can connect over TCP from the benchmark machine. Some local installations configure `root` for socket-only authentication or use a different generated password.

If the database does not exist, create it first:

```sql
CREATE DATABASE benchmark;
```

If you run from a sandboxed terminal and see `operation not permitted`, run the binary from a normal shell or allow local network access.

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
- `.env` is ignored by git; `.env.example` is safe to commit.
- Non-empty databases require confirmation unless `--force` is set.
- Benchmark tables use explicit prefixes.

## Third-party Datasets

This repository does not redistribute Sakila SQL files, TPC-H tools, or generated TPC-H data. See [THIRD_PARTY_LICENSES.md](./THIRD_PARTY_LICENSES.md).

## Development

```bash
make build
make test
make lint
make release-snapshot
go build ./...
```

Release tags build Linux and macOS archives for amd64 and arm64 through GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
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
