package http

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeResponse(w, []byte(swaggerUIHTML))
	})
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		writeResponse(w, swaggerJSON)
	})
	return mux, nil
}

func healthHandler(ping PingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbOK := ping(r.Context()) == nil

		if !dbOK {
			w.WriteHeader(http.StatusServiceUnavailable)
			jsonEncode(w, errorResponse(http.StatusServiceUnavailable, "database unreachable"))
			return
		}

		jsonEncode(w, SuccessResponse(http.StatusOK, map[string]any{
			"db": dbOK,
		}))
	}
}

func writeResponse(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		panic(err)
	}
}

func jsonEncode(w http.ResponseWriter, data interface{}) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		panic(err)
	}
}
