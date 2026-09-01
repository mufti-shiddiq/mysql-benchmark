# Repository Instructions

This repository contains a local MySQL benchmarking CLI written in Go.

## Development

- Keep credentials isolated from result structures and logs.
- Never print passwords, DSNs containing passwords, or `DB_PASSWORD`.
- Keep benchmark-owned tables prefixed with `benchmark_sakila_` or `benchmark_tpch_`.
- Destructive setup must require explicit confirmation unless `--force` is set.
- Run `go test ./...`, `go vet ./...`, and `go build ./...` before considering changes complete.
- Update this document whenever repository behavior, commands, safety rules, or project layout changes.

## Commands

- `make build` builds `./mysql-benchmark`.
- `make test` runs unit tests.
- `make lint` runs `go vet`.
- `make run` starts the CLI through `go run`.
