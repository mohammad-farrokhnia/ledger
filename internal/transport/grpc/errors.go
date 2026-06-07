package grpc

import (
	"errors"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func domainErrorToGRPC(err error) error {
	switch {
	case errors.Is(err, ledger.ErrAccountNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, ledger.ErrInsufficientFunds):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, ledger.ErrCurrencyMismatch):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, ledger.ErrDuplicateTransaction):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, ledger.ErrInvalidAmount):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, ledger.ErrSameAccount):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, ledger.ErrInvalidAccountType):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, ledger.ErrInvalidCurrencyCode):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, "An internal server error occurred")
	}
}
