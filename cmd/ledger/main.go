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

	"github.com/mohammad-farrokhnia/go-ledger/internal/audit"
	"github.com/mohammad-farrokhnia/go-ledger/internal/config"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	_ "github.com/mohammad-farrokhnia/go-ledger/internal/metrics"
	"github.com/mohammad-farrokhnia/go-ledger/internal/store/postgres"
	transportgrpc "github.com/mohammad-farrokhnia/go-ledger/internal/transport/grpc"
	transporthttp "github.com/mohammad-farrokhnia/go-ledger/internal/transport/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	level := slog.LevelInfo
	if err = level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	slog.Info("configuration loaded",
		"grpc_port", cfg.Server.GRPCPort,
		"http_port", cfg.Server.HTTPPort,
		"metrics_port", cfg.Server.MetricsPort,
		"audit_mode", cfg.Audit.Mode,
		"log_level", cfg.Log.Level,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pgStore, err := postgres.New(ctx, cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pgStore.Close()

	slog.Info("connected to postgres")

	var auditor ledger.Auditor
	if cfg.Audit.HookURL != "" {
		auditor = audit.NewWebhookAuditor(cfg.Audit.HookURL, cfg.Audit.Mode)
		slog.Info("audit webhook configured", "url", cfg.Audit.HookURL, "mode", cfg.Audit.Mode)
	} else {
		auditor = &audit.NoOp{}
		slog.Warn("AUDIT_HOOK_URL not set — audit logging disabled")
	}

	svc := ledger.NewService(pgStore, auditor)
	grpcServer := transportgrpc.NewServer(svc)

	gateway, err := transporthttp.NewGateway(
		ctx,
		fmt.Sprintf("localhost:%s", cfg.Server.GRPCPort),
		pgStore.Ping,
	)
	if err != nil {
		return fmt.Errorf("create gateway: %w", err)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: gateway,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("gRPC server starting", "port", cfg.Server.GRPCPort)
		return grpcServer.Start(cfg.Server.GRPCPort)
	})

	g.Go(func() error {
		slog.Info("HTTP gateway starting", "port", cfg.Server.HTTPPort)
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
