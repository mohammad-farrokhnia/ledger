package main

import (
	"context"
	"fmt"
	"log"
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
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
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

	log.Println("connected to postgres")

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
		log.Printf("gRPC server listening on :%s", grpcPort)
		return grpcServer.Start(grpcPort)
	})

	g.Go(func() error {
		log.Printf("HTTP gateway listening on :%s", httpPort)
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()

		log.Println("shutting down...")
		grpcServer.Stop()

		return httpServer.Shutdown(context.Background())
	})

	if err = g.Wait(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	log.Println("shutdown complete")
	return nil
}
