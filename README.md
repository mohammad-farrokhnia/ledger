# go-ledger

A bank-grade financial transaction service built in Go. Designed to be the central financial brain that microservices talk to when they need to move value safely.


## Architecture

```mermaid
graph TD
    A[gRPC Client] -->|ledger.v1.LedgerService| B[gRPC Server :9090]
    C[REST Client] -->|HTTP/JSON| D[HTTP Gateway :8080]
    D -->|reverse proxy| B
    E[Event System] -->|future: Kafka/NATS| F[Messaging Transport]
    F --> G

    B --> G[ledger.Service]
    G -->|Store interface| H[postgres.Store]
    G -->|Auditor interface| I[audit.WebhookAuditor]
    H -->|SELECT FOR UPDATE| J[(PostgreSQL 16)]
    I -->|POST /ingest| K[go-ingestor]
```

### Why hexagonal architecture

go-ledger has three input transports: gRPC, HTTP Gateway, and a future event consumer. Hexagonal architecture isolates the domain from all of them — the service layer never imports gRPC, HTTP, or Postgres packages. Each transport is a plug-in adapter.

### The dependency flow (strictly one-directional, no cycles)

```
cmd/ledger (composition root)
    ↓
internal/ledger (domain: types + Store port + Auditor port + Service)
    ↓ implements Store          ↓ implements Auditor
internal/store/postgres     internal/audit
    ↓
PostgreSQL
```

## How Money Moves (Double-Entry Bookkeeping)

Money is never "changed" — it is only *moved*. Every transaction produces exactly two ledger entries:

```
Transfer $10 from Alice → Bob:

entries table:
  account_id=alice  amount=-10  (DEBIT)
  account_id=bob    amount=+10  (CREDIT)

invariant: SELECT SUM(amount) FROM entries = 0  (always)
```

The `balance` column on the `accounts` table is a denormalized cache updated atomically inside the same DB transaction. If the entries and balances ever disagree, the system has a bug.

## Race Condition Prevention

Two goroutines trying to spend from the same wallet at the same millisecond:

```sql
BEGIN;
-- Locks both account rows. Any other transaction touching these rows WAITS.
SELECT id, type, currency_code, balance
FROM accounts
WHERE id = ANY($1::uuid[])
ORDER BY id          -- ← always same order, prevents deadlock
FOR UPDATE;

-- Check balance in Go against locked data
-- Insert entries and update balances
COMMIT;
-- Locks released. Next waiting transaction proceeds.
```

Without `SELECT FOR UPDATE`: two goroutines read `balance=100`, both think they can spend `$100`, both succeed → balance goes to `-$100`. With it: one waits, reads the updated balance `0`, returns `ErrInsufficientFunds`.

## Idempotency

Clients generate a UUID **before** sending the request. If the network times out:

```
Client → sends request with idempotency_key="abc-123"
Server → processes, commits, returns response
Network → drops the response
Client → retries with same idempotency_key="abc-123"
Server → sees ON CONFLICT DO NOTHING, returns original transaction
Money → moved exactly once
```

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| Database | PostgreSQL 16 |
| Primary transport | gRPC |
| REST transport | gRPC-Gateway (auto-generated) |
| API contract | Protocol Buffers v3 + Buf |
| Migrations | golang-migrate |
| Logging | `log/slog` (JSON) |
| Metrics | Prometheus |
| Config | koanf |

## Quick Start

**Prerequisites:** Go 1.26+, Docker, `buf`, `golang-migrate`, `grpcurl`

```bash
# Clone and set up environment
git clone https://github.com/mohammad-farrokhnia/go-ledger.git
cd go-ledger
cp .env.example .env
cp configs/config.example.yaml configs/config.yaml

# Start Postgres
make up

# Run migrations
make migrate

# Start the service
make run
```

The service starts three ports:
- `:9090` — gRPC
- `:8080` — HTTP gateway + Swagger UI (`/swagger/`) + health (`/healthz`) + metrics (`/metrics`)

## API Examples

### Create wallets

```bash
# SYSTEM wallet — platform revenue account (can go negative, used to mint money)
grpcurl -plaintext -d '{
  "name": "platform-revenue",
  "type": "ACCOUNT_TYPE_SYSTEM",
  "currency_code": "USD"
}' localhost:9090 ledger.v1.LedgerService/CreateWallet

# USER wallet
grpcurl -plaintext -d '{
  "name": "alice",
  "type": "ACCOUNT_TYPE_USER",
  "currency_code": "USD"
}' localhost:9090 ledger.v1.LedgerService/CreateWallet
```

### Top up a user wallet (mint money from SYSTEM)

```bash
grpcurl -plaintext -d '{
  "idempotency_key": "topup-alice-001",
  "from_account_id": "<SYSTEM_ID>",
  "to_account_id": "<ALICE_ID>",
  "amount": 10000,
  "currency_code": "USD"
}' localhost:9090 ledger.v1.LedgerService/CreateTransaction
```

### Check balance via REST

```bash
curl http://localhost:8080/v1/wallets/<ALICE_ID>/balance | jq .
```

### Retry with same idempotency key (safe — returns original)

```bash
# Sending the same request again returns the original transaction, money moves once
grpcurl -plaintext -d '{
  "idempotency_key": "topup-alice-001",
  "from_account_id": "<SYSTEM_ID>",
  "to_account_id": "<ALICE_ID>",
  "amount": 10000,
  "currency_code": "USD"
}' localhost:9090 ledger.v1.LedgerService/CreateTransaction
```

### Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Prometheus metrics

```bash
curl http://localhost:8080/metrics | grep ledger
```

## Developer Commands

```bash
make up            # start Postgres + pgAdmin (localhost:5050)
make down          # stop containers, keep data
make down-volumes  # stop containers, delete all data
make migrate       # run migrations on dev DB
make migrate-test  # run migrations on test DB
make gen           # regenerate proto → Go + Swagger
make lint          # run golangci-lint
make test          # run all tests with -race flag
make run           # run the service (reads .env)
make db            # open psql shell
make help          # list all targets
```

## Configuration

Copy `.env.example` → `.env`. Never commit `.env`.

| Variable | Default | Description |
|---|---|---|
| `DB_DSN` | required | Postgres connection string |
| `GRPC_PORT` | `9090` | gRPC server port |
| `HTTP_PORT` | `8080` | HTTP gateway port |
| `METRICS_PORT` | `9091` | Prometheus metrics port |
| `AUDIT_MODE` | `async` | `sync` blocks, `async` fire-and-forget |
| `AUDIT_HOOK_URL` | — | go-ingestor endpoint, disabled if empty |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `COMPOSE_PROJECT_NAME` | `docker` | the group name of continer in docker |

## Running Tests

```bash
# Ensure test DB is migrated
make migrate-test

# Run all tests including concurrency tests
make test

# The key tests to look at:
# TestCreateTransaction_ConcurrentDeductions  — 50 goroutines, proves no race
# TestCreateTransaction_ConcurrentContention  — 100 goroutines, proves ErrInsufficientFunds
# TestCreateTransaction_Idempotency           — same key twice, money moves once
```

## Building the Docker Image

```bash
docker build -f deployments/docker/Dockerfile -t go-ledger:latest .

# Run with docker compose (includes Postgres)
docker compose -f deployments/docker/docker-compose.yml up
```

## Error Reference

| Error | gRPC Code | Meaning |
|---|---|---|
| `account not found` | `NOT_FOUND` | Account ID does not exist |
| `transaction not found` | `NOT_FOUND` | Transaction ID does not exist |
| `insufficient funds` | `FAILED_PRECONDITION` | Sender balance too low |
| `currency mismatch` | `INVALID_ARGUMENT` | Accounts have different currencies |
| `transaction with this idempotency key already exists` | `ALREADY_EXISTS` | Duplicate — fetch original |
| `amount must be greater than zero` | `INVALID_ARGUMENT` | Invalid amount |
| `source and destination accounts must be different` | `INVALID_ARGUMENT` | Same account |
| `an internal error occurred` | `INTERNAL` | Server error (details logged server-side) |

## Project Structure

```
api/proto/ledger/v1/   proto definition + generated Go code
cmd/ledger/            entrypoint — composition root only
configs/               config.yaml template
deployments/docker/    Dockerfile + docker-compose
internal/
  audit/               Auditor implementations (NoOp, Webhook)
  config/              koanf config loader
  ledger/              domain: types, errors, Store/Auditor ports, Service
  metrics/             Prometheus metric definitions
  store/postgres/      Postgres implementation of ledger.Store
  transport/grpc/      gRPC server, handler, interceptors, mappers
  transport/http/      gRPC-Gateway, health check, Swagger UI
migrations/            SQL migration files (up + down)
```
