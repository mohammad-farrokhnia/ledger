package http

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	ledgerv1 "github.com/mohammad-farrokhnia/ledger/api/proto/ledger/v1"
)

//go:embed swagger.json
var swaggerJSON []byte

type PingFunc func(ctx context.Context) error

func NewGateway(ctx context.Context, grpcAddr string, ping PingFunc) (http.Handler, error) {
	gwMux := runtime.NewServeMux(
		runtime.WithErrorHandler(CustomErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   false,
			},
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := ledgerv1.RegisterLedgerServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("gateway: register handler: %w", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/v1/", gwMux)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthHandler(ping))

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write(swaggerJSON); err != nil {
			http.Error(w, "failed to serve swagger spec", http.StatusInternalServerError)
		}
	})

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))

	return mux, nil
}
