package http

import (
	"context"
	"encoding/json"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
)

var swaggerJSON []byte

type PingFunc func(ctx context.Context) error

func NewGateway(ctx context.Context, grpcAddr string, ping PingFunc) (http.Handler, error) {
	gwMux := runtime.NewServeMux()

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
		writeResponse(w, swaggerJSON)
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeResponse(w, []byte(swaggerUIHTML))
	})

	return mux, nil
}

func healthHandler(ping PingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
				"status": "degraded",
				"reason": "database unreachable",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	}
}

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <title>go-ledger API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css"
    href="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/swagger.json",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  })
</script>
</body>
</html>`
func writeResponse(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		panic(err)
	}
}

