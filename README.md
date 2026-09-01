# MySQL Benchmark

[![CI](https://github.com/mufti-shiddiq/mysql-benchmark/actions/workflows/ci.yml/badge.svg)](https://github.com/mufti-shiddiq/mysql-benchmark/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/mufti-shiddiq/mysql-benchmark)](https://github.com/mufti-shiddiq/mysql-benchmark/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Benchmark MySQL from the same machine your application uses.

`mysql-benchmark` is a small CLI for comparing MySQL providers, plans, or regions from an Ubuntu VPS or any other app server. It measures the parts that usually matter in real deployments: connection latency, simple queries, joins, aggregations, writes, transactions, and analytical queries.

It runs locally. It does not send results anywhere.

## When To Use This

Use this when you want to answer questions like:

- Is Provider A faster than Provider B from my VPS?
- How much latency do I pay when the database is in another region?
- Are joins or analytical queries noticeably slower on this MySQL plan?
- Is write/transaction latency acceptable from my app server?

For fair comparisons, run every test from the same machine. A benchmark from your laptop is not comparable to a benchmark from your production VPS.

```text
Same Ubuntu VPS
  -> MySQL Provider A
  -> MySQL Provider B
```

## Install

Install the latest release binary:

```bash
curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh | sh
```

Install a specific version:

```bash
MYSQL_BENCHMARK_VERSION=v0.1.0 sh -c "$(curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh)"
```

Then run:

```bash
mysql-benchmark
```

If `/usr/local/bin` requires sudo, the installer will ask for it. To install without sudo:

```bash
INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh)"
```

Make sure `$HOME/.local/bin` is in your `PATH`.

## Requirements

You need:

- A reachable MySQL-compatible database.
- An existing database/schema. The tool does not create databases automatically.
- A MySQL user allowed to create, insert, update, delete, and drop benchmark-owned tables.

You only need Go if you want to build from source.

## First Run

Interactive mode is the easiest way to start:

```bash
mysql-benchmark
```

The CLI asks for:

- host
- port
- database
- username
- password
- benchmark mode

Passwords are not printed and are not included in JSON output.

For local MySQL, use TCP explicitly:

```bash
mysql-benchmark --host 127.0.0.1 --port 3306 --database benchmark --user root
```

If the database does not exist yet, create it first:

```sql
CREATE DATABASE benchmark;
```

## Configuration

You can configure the CLI in three ways:

- interactive prompts
- `.env` file
- shell environment variables
- CLI flags

Precedence is:

```text
defaults < .env file < shell environment < CLI flags
```

Create a local `.env`:

```bash
cp .env.example .env
```

Example:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=benchmark
DB_USER=root
```

The CLI reads `.env` by default. You can also pass another file:

```bash
mysql-benchmark --env-file production.env --mode sakila
```

For automation, provide the password through `DB_PASSWORD`:

```bash
DB_PASSWORD='secret' mysql-benchmark \
  --host db.example.com \
  --database benchmark \
  --user benchmark \
  --mode both
```

Avoid passing passwords as CLI arguments because they can appear in shell history or process listings.

## Common Commands

Run Sakila-style workload:

```bash
mysql-benchmark --mode sakila
```

Run TPC-H-inspired workload:

```bash
mysql-benchmark --mode tpch
```

Run both:

```bash
mysql-benchmark --mode both
```

Write JSON to stdout:

```bash
mysql-benchmark --output json
```

Write JSON to a file:

```bash
mysql-benchmark --output results.json
```

Example JSON shape:

```json
{
  "database": {
    "host": "db.example.com",
    "port": 3306,
    "database": "benchmark",
    "mysql_version": "8.0.43"
  },
  "benchmark": "sakila",
  "scale_factor": 1,
  "configuration": {
    "warmup": 5,
    "iterations": 30
  },
  "results": [
    {
      "name": "select_1_latency",
      "p50_ms": 3.12,
      "p95_ms": 4.21,
      "p99_ms": 5.82
    }
  ]
}
```

Use fewer iterations for a quick smoke test:

```bash
mysql-benchmark --warmup 1 --iterations 3
```

## Flags

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

## Safety

The tool uses benchmark-owned table prefixes:

- `benchmark_sakila_`
- `benchmark_tpch_`

If the selected database is not empty, the CLI asks for confirmation before preparing benchmark data. The default answer is no.

Use `--force` only in automation when you understand that benchmark-owned tables may be dropped and recreated:

```bash
mysql-benchmark --mode both --force
```

The tool should be run against a dedicated benchmark database, not an application production schema.

## Benchmark Modes

### Sakila-style

This workload creates synthetic relational tables inspired by the shape of Sakila. It tests:

- `SELECT 1` latency
- indexed reads
- two-table joins
- multi-table joins
- complex joins
- `COUNT`, `SUM`, `AVG`, and `GROUP BY`
- sorting with `ORDER BY` and `LIMIT`
- subqueries
- insert, update, delete, and transaction latency

### TPC-H-inspired

This workload creates synthetic analytical tables using the familiar TPC-H table families:

- region
- nation
- supplier
- customer
- part
- partsupp
- orders
- lineitem

It runs query cases named Q01 through Q22 where practical for MySQL. This is a TPC-H-inspired workload, not an official TPC-H compliant result.

## Methodology

Each benchmark case runs:

1. warmup iterations
2. measured iterations
3. metric calculation

Warmup timings are discarded. Setup time is separate from query timing.

Default settings:

```text
warmup: 5
iterations: 30
```

Reported metrics:

- min
- avg
- p50
- p95
- p99
- max

All timings are shown in milliseconds.

`SELECT 1 latency` means client/server round trip plus MySQL processing overhead. It is not pure TCP network RTT.

## Reading Results

Use the results comparatively. They are affected by:

- distance between app server and database
- MySQL version and configuration
- CPU, RAM, and storage
- indexes and query plan choices
- warm cache vs cold cache
- concurrent workload on the database
- network jitter

Run the same benchmark more than once if you need confidence. For provider comparisons, run from the same VPS and change only the database target.

For more detail, see [docs/methodology.md](./docs/methodology.md).

## Troubleshooting

Access denied:

```text
access denied
```

Check username, password, and whether the MySQL user can connect from the benchmark host. Some local MySQL installs configure `root` for socket-only authentication.

Database missing:

```text
database "benchmark" does not exist
```

Create the database manually, then rerun the benchmark.

Connection refused or timeout:

```text
unable to reach MySQL
```

Check host, port, firewall rules, bind address, and whether MySQL is running.

Sandboxed terminal:

```text
operation not permitted
```

Run the binary from a normal shell or allow local network access.

## Third-party Datasets

This repository does not redistribute Oracle Sakila SQL files, TPC-H tools, or generated TPC-H data. The CLI generates benchmark-owned synthetic data locally.

See [THIRD_PARTY_LICENSES.md](./THIRD_PARTY_LICENSES.md).

## Build From Source

```bash
git clone https://github.com/mufti-shiddiq/mysql-benchmark.git
cd mysql-benchmark
go build -o mysql-benchmark ./cmd/mysql-benchmark
./mysql-benchmark
```

## Development

Contributions are welcome. Start with [CONTRIBUTING.md](./CONTRIBUTING.md).

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

Maintainer release notes are in [docs/release.md](./docs/release.md).

Security issues should be reported privately. See [SECURITY.md](./SECURITY.md).

Repository metadata can be configured with GitHub CLI after authentication:

```bash
gh auth login
scripts/configure-github.sh
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
