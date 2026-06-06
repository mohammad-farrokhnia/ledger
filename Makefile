.PHONY: lint test build run up down migrate gen

# Run the linter
lint:
	golangci-lint run ./...

# Run all tests with race detector enabled
test:
	go test -race -count=1 ./...

# Build the binary
build:
	go build -o bin/ledger ./cmd/ledger

# Run the service locally
run:
	go run ./cmd/ledger

# Start local infrastructure (postgres)
up:
	docker compose -f deployments/docker/docker-compose.yml up -d

# Stop local infrastructure
down:
	docker compose -f deployments/docker/docker-compose.yml down

# Run database migrations (up)
migrate:
	migrate -path migrations -database "$(DB_DSN)" up

# Generate proto files
gen:
	buf generate