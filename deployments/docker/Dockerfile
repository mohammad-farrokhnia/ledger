FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o ledger \
    ./cmd/ledger

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /build/ledger .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/configs/config.example.yaml ./configs/config.yaml

EXPOSE 9090 8080 9091

ENTRYPOINT ["/app/ledger"]
