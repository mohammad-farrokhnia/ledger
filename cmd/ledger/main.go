package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"github.com/mohammad-farrokhnia/go-ledger/internal/store/postgres"
	transportgrpc "github.com/mohammad-farrokhnia/go-ledger/internal/transport/grpc"
	transporthttp "github.com/mohammad-farrokhnia/go-ledger/internal/transport/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return fmt.Errorf("DB_DSN environment variable is required")
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pgStore, err := postgres.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pgStore.Close()

	slog.Info("connected to postgres")

	svc := ledger.NewService(pgStore)
	grpcServer := transportgrpc.NewServer(svc)

	gateway, err := transporthttp.NewGateway(ctx, fmt.Sprintf("localhost:%s", grpcPort))
	if err != nil {
		return fmt.Errorf("create gateway: %w", err)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: gateway,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("gRPC server starting", "port", grpcPort)
		return grpcServer.Start(grpcPort)
	})

	g.Go(func() error {
		slog.Info("HTTP gateway starting", "port", httpPort)
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		slog.Info("shutdown signal received")
		grpcServer.Stop()
		return httpServer.Shutdown(context.Background())
	})

	if err = g.Wait(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	slog.Info("shutdown complete")
	return nil
}
