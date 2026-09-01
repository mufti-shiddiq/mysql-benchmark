# Release Process

This repository publishes prebuilt binaries through GitHub Releases.

## Prepare

1. Ensure the changelog has an entry for the version.
2. Run local verification:

```bash
go test ./...
go vet ./...
go build ./...
make release-snapshot
```

3. Confirm release archives exist:

```bash
ls dist/*.tar.gz
```

## Tag

Use semantic version tags:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Pushing a `v*` tag starts `.github/workflows/release.yml`.

## Verify

After the GitHub Action completes, confirm the release includes:

- `mysql-benchmark-linux-amd64.tar.gz`
- `mysql-benchmark-linux-arm64.tar.gz`
- `mysql-benchmark-darwin-amd64.tar.gz`
- `mysql-benchmark-darwin-arm64.tar.gz`
- `checksums.txt`

Then test the installer:

```bash
MYSQL_BENCHMARK_VERSION=v0.1.0 sh -c "$(curl -fsSL https://raw.githubusercontent.com/mufti-shiddiq/mysql-benchmark/main/scripts/install.sh)"
mysql-benchmark --version
```

## GitHub Repository Metadata

Recommended About description:

```text
Local CLI to benchmark MySQL latency, joins, writes, transactions, and TPC-H-inspired workloads from your app server.
```

Recommended topics:

```text
mysql
benchmark
database
database-benchmark
performance
latency
sql
cli
golang
go
mysql-benchmark
tpch
sakila
devops
vps
rds
mariadb
observability
```

Use `scripts/configure-github.sh` to replace the repository topic set with that list, keeping the repo within GitHub's 20-topic limit.

After authenticating GitHub CLI, repository metadata and branch protection can be configured with:

```bash
scripts/configure-github.sh
```

If branch protection setup fails, configure it manually in GitHub:

- protect `main`
- require CI status checks before merge
- require one approving review
- require linear history
- disable force pushes
- disable branch deletion
