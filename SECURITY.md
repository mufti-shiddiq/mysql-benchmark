# Security Policy

## Supported Versions

Security fixes are planned for the latest released minor version.

| Version | Supported |
| ------- | --------- |
| 0.1.x   | Yes       |

## Reporting a Vulnerability

Please do not open a public issue for security vulnerabilities.

Report privately through GitHub Security Advisories for this repository:

<https://github.com/mufti-shiddiq/mysql-benchmark/security/advisories/new>

If that is unavailable, contact the maintainer through the GitHub profile for `mufti-shiddiq`.

## Scope

Security-sensitive areas include:

- password handling
- DSN construction
- JSON/text output redaction
- benchmark table cleanup
- `.env` handling
- installer and release artifacts

The CLI must never print passwords, store passwords in result files, or log full credential-bearing connection strings.
