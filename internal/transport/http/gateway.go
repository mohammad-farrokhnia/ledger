package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
)

func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := ledgerv1.RegisterLedgerServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("gateway: register handler: %w", err)
	}

	return mux, nil
}
