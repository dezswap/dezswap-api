# Contributing to dezswap-api

## Branching

- Base all changes off `main`.
- Use descriptive branch names: `feat/short-description`, `fix/short-description`.

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/). Each commit message must have a type prefix:

| Prefix | When to use |
|---|---|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `refactor:` | Code change that is not a feature or fix |
| `perf:` | Performance improvement |
| `docs:` | Documentation only |
| `test:` | Adding or updating tests |
| `chore:` | Maintenance (deps, tooling, config) |
| `ci:` | CI/CD changes |

`docs:`, `test:`, `chore:`, `ci:`, and `build:` commits are excluded from the changelog automatically.

## Pull Requests

1. Keep PRs focused — one concern per PR.
2. Fill in the [PR template](.github/pull_request_template.md).
3. All CI checks (lint, test, build) must pass before merging.

## Development Setup

```bash
# Install dependencies
go mod download

# Config setup
cp config.example.yml config.yml

# Start DB, Redis, and app (indexer by default)
make up
APP_TYPE=api make up   # or start the API server
```

> The database schema depends on migrations from [cosmwasm-etl](https://github.com/dezswap/cosmwasm-etl). Run those first.

```bash
make api-migrate-up
```

## Code Quality

```bash
make fmt        # format
make lint       # golangci-lint
make test       # unit tests (-short)
make test-race  # tests with race detector
```

All tests must pass with `make test-race`. Integration tests requiring a live DB are gated by the `-short` flag and run in CI only.

## Generating Code

```bash
make generate   # regenerates ERC20 bindings via abigen
```

Run this after modifying any ABI files under `pkg/erc20/`.
