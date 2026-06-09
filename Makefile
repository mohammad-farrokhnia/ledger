include .env
export

.PHONY: lint test build run \
        up down down-volumes logs db \
        migrate migrate-down migrate-create \
        gen help

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## test: run all tests with race detector
test:
	go test -race -count=1 -timeout 120s ./...

## test-verbose: run tests with full output
test-verbose:
	go test -race -count=1 -timeout 120s -v ./...

## build: compile the binary to bin/ledger
build:
	go build -o bin/ledger ./cmd/ledger

## run: run the service locally
run:
	go run ./cmd/ledger


## up: start containers
up:
	docker compose -f deployments/docker/docker-compose.yml up -d
	@echo "Postgres ready at localhost:5432"
	@echo "pgAdmin ready at http://localhost:5050 (admin@ledger.dev / admin)"

## down: stop containers (data is preserved)
down:
	docker compose -f deployments/docker/docker-compose.yml down

## down-volumes: stop containers AND delete all data (clean slate)
down-volumes:
	docker compose -f deployments/docker/docker-compose.yml down -v

## logs: tail container logs
logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f

## db: open a psql shell directly inside the postgres container
db:
	docker exec -it ledger-postgres psql -U postgres -d ledger


## migrate: run all pending migrations (up)
migrate:
	migrate -path migrations -database "$(DB_DSN)" up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path migrations -database "$(DB_DSN)" down 1

## migrate-create name=<migration_name>: create a new migration file pair
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-test:
	migrate -path migrations -database "$(DB_DSN_TEST)" up

migrate-down-test:
	migrate -path migrations -database "$(DB_DSN_TEST)" down 1


## gen: generate Go code and Swagger from .proto files
gen:
	buf generate
	cp api/openapi/ledger.swagger.json internal/transport/http/swagger.json


## help: print this help message
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'