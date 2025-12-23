# go-ledger 🏛️

[![Go Report Card](https://goreportcard.com/badge/github.com/mohammad-farrokhnia/go-ledger/actions)](https://goreportcard.com/report/github.com/mohammad-farrokhnia/go-ledger)
[![Build Status](https://github.com/mohammad-farrokhnia/go-ledger/actions/workflows/ci.yml/badge.svg)](https://github.com/mohammad-farrokhnia/go-ledger/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**go-ledger** is a high-consistency, transactional financial service designed to handle money movement with strict atomicity and correctness. 

It implements the **Double-Entry Bookkeeping** principle, ensuring that money is never created or destroyed without a traceable source. It is designed to be protocol-agnostic, serving as the financial backbone for systems requiring auditability and precision.

## 🚀 Key Features

* **Atomic Transactions:** Uses `SELECT FOR UPDATE` and database-level locking to prevent race conditions and ensure ACID compliance.
* **Double-Entry Core:** Every transaction writes two immutable entries (Credit/Debit), ensuring `Sum(Entries) = 0`.
* **Multi-Protocol Access:** Native **gRPC** support with auto-generated **REST/JSON** gateway (OpenAPI/Swagger).
* **Idempotency:** Native support for idempotency keys to safely retry network requests without double-spending.
* **Multi-Currency:** Built-in support for multiple ISO-4217 currencies within the same system.
* **Audit Hooks:** Decoupled interface for async/sync shipping of transaction logs to external ingestors (e.g., for data warehousing).

## 🏗️ Architecture

This project follows a **Clean/Hexagonal Architecture**:

* **`api/`**: Protocol definitions (Protobuf/gRPC).
* **`internal/ledger`**: Pure domain logic. Knows nothing about HTTP or SQL.
* **`internal/store`**: The Data Layer. Handles SQL transactions and locking logic.
* **`internal/transport`**: Adapters (Ports) for gRPC, HTTP Gateway, and Event Consumers.

## 🛠️ Tech Stack

* **Language:** Go 1.22+
* **Database:** PostgreSQL 15+
* **Communication:** gRPC / grpc-gateway
* **Observability:** Prometheus Metrics, Structured Logging (Zap)
* **Tooling:** Buf (Proto management), GolangCI-Lint, Docker

## ⚡ Quick Start

### Prerequisites
* Go 1.22+
* Docker & Docker Compose
* Make

### Running Locally

1.  **Start Infrastructure:**
    ```bash
    make up  # Starts Postgres and local dev tools
    ```

2.  **Run Migrations:**
    ```bash
    make migrate
    ```

3.  **Run the Service:**
    ```bash
    go run cmd/ledger/main.go
    ```

4.  **Test the API:**
    Access Swagger UI at `http://localhost:8080/swagger/` or use the REST endpoint:
    ```bash
    curl -X POST http://localhost:8080/v1/wallets -d '{"currency":"USD", "type":"USER"}'
    ```

## 🧪 Development Standards

We enforce strict linting and testing rules. Before pushing:

```bash
make lint   # Runs golangci-lint with strict settings
make test   # Runs unit tests with race detection