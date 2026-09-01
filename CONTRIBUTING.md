# Contributing

Thanks for taking the time to improve `mysql-benchmark`.

This project is intentionally small: a local Go CLI that benchmarks MySQL from the same machine your application uses. Contributions should keep that scope clear and avoid adding hosted services, telemetry, Docker requirements, or support for other database engines unless the roadmap changes.

## Development Setup

```bash
git clone https://github.com/mufti-shiddiq/mysql-benchmark.git
cd mysql-benchmark
go test ./...
go vet ./...
go build ./...
```

Build the runnable binary:

```bash
go build -o mysql-benchmark ./cmd/mysql-benchmark
```

## Before Opening a Pull Request

Run:

```bash
go test ./...
go vet ./...
go build ./...
```

For release archive changes, also run:

```bash
make release-snapshot
```

## Commit Messages

Use Conventional Commits:

```text
feat: add new benchmark case
fix: correct mysql connection handling
docs: improve install guide
test: add metrics coverage
chore: update ci workflow
```

## Safety Rules

- Never print passwords.
- Never include credentials or full DSNs in logs, errors, or result files.
- Keep benchmark-owned tables prefixed with `benchmark_sakila_` or `benchmark_tpch_`.
- Require explicit confirmation for destructive setup unless `--force` is used.
- Prefer a dedicated benchmark database over an application schema.

## Pull Request Checklist

- The change is focused and matches the project scope.
- Tests were added or updated when behavior changed.
- Documentation was updated for user-visible behavior.
- `AGENTS.md` was updated when commands, layout, safety rules, or project behavior changed.
- `go test ./...`, `go vet ./...`, and `go build ./...` pass.
