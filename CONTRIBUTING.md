# Contributing to tlog

Thank you for your interest in contributing! This document covers how to set up the project locally, run tests, and submit changes.

---

## Prerequisites

- Go 1.26.3 or later
- [`golangci-lint`](https://golangci-lint.run/usage/install/) for linting

---

## Development Setup

```bash
git clone https://github.com/tidjee-dev/tlog.git
cd tlog
go mod tidy
```

---

## Common Commands

| Command      | Description                      |
| ------------ | -------------------------------- |
| `make build` | Compile the project              |
| `make test`  | Run all tests with race detector |
| `make lint`  | Run golangci-lint                |
| `make bench` | Run all benchmarks               |
| `make tidy`  | Tidy go.mod / go.sum             |
| `make clean` | Remove build artifacts           |

---

## Pull Request Guidelines

1. **Fork** the repository and create your branch from `main`.
2. **Write tests** for any new functionality.
3. **Run `make lint` and `make test`** — both must pass before opening a PR.
4. **Commit messages** should be concise and written in the imperative mood (e.g., `Add JSON formatter`).
5. **One concern per PR** — keep changes focused.

---

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`).
- All exported symbols must have a godoc comment.
- No `reflect` in hot logging paths — see [PLAN.md](PLAN.md) for rationale.
- Avoid `fmt.Sprintf` in hot paths; use manual string building or buffer writes.

---

## Reporting Issues

Please open a GitHub issue with a clear description, minimal reproduction case, and Go version.
