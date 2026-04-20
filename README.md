# dezswap-api

Backend API and indexer service for [Dezswap](https://dezswap.io).

## Overview

This repository contains two independent binaries that share a PostgreSQL database:

- **`indexer`** — background worker that periodically syncs token metadata, liquidity pool states, and verified token status from blockchains via gRPC.
- **`api`** — Gin HTTP server exposing REST endpoints for the Dezswap frontend, plus CoinGecko and CoinMarketCap compatibility endpoints.

## Prerequisites

- Go 1.23+
- Docker (for PostgreSQL and Redis)
- [golangci-lint](https://golangci-lint.run/) (for linting)

## Getting Started

```bash
# 1. Clone
git clone https://github.com/dezswap/dezswap-api.git
cd dezswap-api

# 2. Configure
cp config.example.yml config.yml
# Edit config.yml with your chain and DB settings

# 4. Build and run
make api       # builds ./main for the API server
make indexer   # builds ./main for the indexer
```

## Configuration

Configuration is loaded from `config.yml` by default. Environment variables are supported using the `APP_` prefix with dots replaced by underscores (e.g., `APP_API_SERVER_PORT=8000`).

Key sections in `config.yml`:

| Section | Description |
|---|---|
| `indexer` | Chain ID, gRPC node endpoint, EVM RPC, source DB |
| `api.server` | Host, port, CORS origins, Swagger toggle |
| `api.db` | PostgreSQL connection |
| `api.cache` | Redis or in-memory cache |
| `log` | Log level and format |
| `sentry` | Optional Sentry DSN for error tracking |

See [`config.example.yml`](config.example.yml) for the full reference.

### Database Migrations

The database schema depends on migrations managed by [cosmwasm-etl](https://github.com/dezswap/cosmwasm-etl). Run those migrations first before applying the API-specific ones below.

```bash
make api-migrate-up
make api-migrate-down
make api-generate-migration
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](LICENSE).
