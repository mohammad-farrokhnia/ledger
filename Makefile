.PHONY: lint test gen

# Run the linter
lint:
	golangci-lint run

# Run tests with race detection
test:
	go test -race -v ./...

# Generate Proto & Swagger (Requires buf installed)
gen:
	buf generate