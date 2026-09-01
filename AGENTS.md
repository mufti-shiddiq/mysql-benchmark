# Repository Instructions

This repository contains a local MySQL benchmarking CLI written in Go.

## Development

- Keep credentials isolated from result structures and logs.
- Never print passwords, DSNs containing passwords, or `DB_PASSWORD`.
- Keep benchmark-owned tables prefixed with `benchmark_sakila_` or `benchmark_tpch_`.
- Load local configuration with `.env` support using precedence `defaults < .env file < shell environment < CLI flags`.
- Keep connection errors human-readable and avoid exposing raw credential-bearing connection details.
- Use a single MySQL connection charset value such as `utf8mb4`; comma-separated charset fallbacks are escaped by `go-sql-driver/mysql` DSN formatting and can break connection setup.
- Destructive setup must require explicit confirmation unless `--force` is set.
- Run `go test ./...`, `go vet ./...`, and `go build ./...` before considering changes complete.
- Release binaries are produced by `.github/workflows/release.yml` for Linux/macOS amd64/arm64 when a `v*` tag is pushed.
- Keep `scripts/install.sh` compatible with POSIX `sh`, Ubuntu VPS defaults, `curl` or `wget`, and user-writable `INSTALL_DIR`.
- Update this document whenever repository behavior, commands, safety rules, or project layout changes.

## Commands

- `make build` builds `./mysql-benchmark`.
- `make test` runs unit tests.
- `make lint` runs `go vet`.
- `make run` starts the CLI through `go run`.
- `make release-snapshot` builds Linux amd64/arm64 release archives locally.
