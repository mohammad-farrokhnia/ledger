package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

func domainErrorToGRPC(err error) error {
	var valErr *ledger.ValidationError
	if errors.As(err, &valErr) {
		return status.Error(codes.InvalidArgument, valErr.Error())
	}

	switch {
	case errors.Is(err, ledger.ErrAccountNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, ledger.ErrTransactionNotFound):
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
		return status.Error(codes.Internal, "an internal error occurred")
	}
}

func logHandlerError(ctx context.Context, method string, err error) {
	var valErr *ledger.ValidationError
	switch {
	case errors.As(err, &valErr),
		errors.Is(err, ledger.ErrAccountNotFound),
		errors.Is(err, ledger.ErrTransactionNotFound),
		errors.Is(err, ledger.ErrInsufficientFunds),
		errors.Is(err, ledger.ErrCurrencyMismatch),
		errors.Is(err, ledger.ErrDuplicateTransaction),
		errors.Is(err, ledger.ErrInvalidAmount),
		errors.Is(err, ledger.ErrSameAccount),
		errors.Is(err, ledger.ErrInvalidAccountType),
		errors.Is(err, ledger.ErrInvalidCurrencyCode):
		slog.WarnContext(ctx, method+" rejected", "reason", err.Error())
	default:
		slog.ErrorContext(ctx, method+" failed", "error", err)
	}
}