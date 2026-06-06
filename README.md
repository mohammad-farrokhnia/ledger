# go-ledger

A high-consistency, bank-grade financial transaction service built in Go.

> Status: 🚧 Under active development

## Architecture

Double-entry bookkeeping + ACID transactions + idempotency keys.

## Tech Stack

- **Go 1.22+**
- **PostgreSQL** (row-level locking via `SELECT FOR UPDATE`)
- **gRPC + gRPC-Gateway** (serves both gRPC and REST/JSON)
- **Prometheus** (metrics)
- **Zap** (structured logging)

## How to Run

```bash
make up      # start postgres
make migrate # run DB migrations
make run     # start the service
```

## Documentation

See `api/openapi/` for the Swagger spec once generated.