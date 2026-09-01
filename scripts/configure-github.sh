#!/usr/bin/env sh
set -eu

REPO="${1:-mufti-shiddiq/mysql-benchmark}"
DESCRIPTION="Local CLI to benchmark MySQL latency, joins, writes, transactions, and TPC-H-inspired workloads from your app server."
TOPICS_JSON='{"names":["mysql","benchmark","database","database-benchmark","performance","latency","sql","cli","golang","go","mysql-benchmark","tpch","sakila","devops","vps","rds","mariadb","observability"]}'

if ! command -v gh >/dev/null 2>&1; then
	echo "missing required command: gh" >&2
	exit 1
fi

gh auth status >/dev/null

gh repo edit "$REPO" \
	--description "$DESCRIPTION" \
	--homepage "https://github.com/$REPO" \
	--enable-issues \
	--enable-secret-scanning \
	--enable-secret-scanning-push-protection \
	--enable-squash-merge \
	--enable-merge-commit=false \
	--enable-rebase-merge \
	--delete-branch-on-merge \
	--enable-wiki=false

printf '%s' "$TOPICS_JSON" | gh api \
	--method PUT \
	-H "Accept: application/vnd.github+json" \
	"/repos/$REPO/topics" \
	--input -

PROTECTION_BODY='{
  "required_status_checks": {
    "strict": true,
    "contexts": ["test"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}'

printf '%s' "$PROTECTION_BODY" | gh api \
	--method PUT \
	-H "Accept: application/vnd.github+json" \
	"/repos/$REPO/branches/main/protection" \
	--input -

echo "Configured GitHub metadata and branch protection for $REPO"
