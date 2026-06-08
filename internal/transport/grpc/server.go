package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type Server struct {
	grpc *grpc.Server
}

func NewServer(svc ledger.Servicer) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor,
			LoggingInterceptor,
		),
	)

	ledgerv1.RegisterLedgerServiceServer(grpcServer, NewHandler(svc))
	reflection.Register(grpcServer)

	return &Server{grpc: grpcServer}
}

func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("grpc: listen on port %s: %w", port, err)
	}

	return s.grpc.Serve(lis)
}

func (s *Server) Stop() {
	s.grpc.GracefulStop()
}
