package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"github.com/mohammad-farrokhnia/go-ledger/internal/metrics"
)

type Handler struct {
	ledgerv1.UnimplementedLedgerServiceServer
	svc ledger.Servicer
}

func NewHandler(svc ledger.Servicer) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateWallet(ctx context.Context, req *ledgerv1.CreateWalletRequest) (*ledgerv1.CreateWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	acc, err := h.svc.CreateAccount(ctx, req.Name, protoAccountTypeToDomain(req.Type), req.CurrencyCode)
	if err != nil {
		logHandlerError(ctx, "CreateWallet", err)
		return nil, domainErrorToGRPC(err)
	}

	slog.InfoContext(ctx, "wallet created", "account_id", acc.ID, "type", acc.Type)
	return &ledgerv1.CreateWalletResponse{Wallet: accountToProto(acc)}, nil
}

func (h *Handler) GetBalance(ctx context.Context, req *ledgerv1.GetBalanceRequest) (*ledgerv1.GetBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	acc, err := h.svc.GetAccount(ctx, req.WalletId)
	if err != nil {
		slog.ErrorContext(ctx, "GetBalance failed", "error", err, "wallet_id", req.WalletId)
		return nil, domainErrorToGRPC(err)
	}

	return &ledgerv1.GetBalanceResponse{
		WalletId:     acc.ID,
		Balance:      acc.Balance,
		CurrencyCode: acc.CurrencyCode,
	}, nil
}

func (h *Handler) CreateTransaction(ctx context.Context, req *ledgerv1.CreateTransactionRequest) (*ledgerv1.CreateTransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	start := time.Now()

	tx, err := h.svc.CreateTransaction(ctx, ledger.CreateTransactionInput{
		IdempotencyKey: req.IdempotencyKey,
		FromAccountID:  req.FromAccountId,
		ToAccountID:    req.ToAccountId,
		Amount:         req.Amount,
		CurrencyCode:   req.CurrencyCode,
	})

	duration := time.Since(start).Seconds()
	metrics.TransactionDuration.Observe(duration)

	if err != nil {
		errorType := classifyError(err)
		metrics.TransactionsTotal.WithLabelValues("fail", errorType).Inc()
		slog.ErrorContext(ctx, "CreateTransaction failed",
			"error", err,
			"error_type", errorType,
			"idempotency_key", req.IdempotencyKey,
		)
		return nil, domainErrorToGRPC(err)
	}

	metrics.TransactionsTotal.WithLabelValues("success", "").Inc()

	slog.InfoContext(ctx, "transaction created",
		"transaction_id", tx.ID,
		"from", tx.FromAccountID,
		"to", tx.ToAccountID,
		"amount", tx.Amount,
		"currency", tx.CurrencyCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &ledgerv1.CreateTransactionResponse{Transaction: transactionToProto(tx)}, nil
}

func (h *Handler) GetWalletHistory(ctx context.Context, req *ledgerv1.GetWalletHistoryRequest) (*ledgerv1.GetWalletHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	var offset int32
	if req.PageToken != "" {
		n, err := strconv.Atoi(req.PageToken)
		if err != nil || n < 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		offset = int32(n)
	}

	entries, err := h.svc.GetWalletHistory(ctx, req.WalletId, req.PageSize, offset)
	if err != nil {
		slog.ErrorContext(ctx, "GetWalletHistory failed", "error", err, "wallet_id", req.WalletId)
		return nil, domainErrorToGRPC(err)
	}

	protoEntries := make([]*ledgerv1.Entry, len(entries))
	for i, e := range entries {
		protoEntries[i] = entryToProto(e)
	}

	var nextToken string
	if req.PageSize > 0 && int32(len(entries)) == req.PageSize {
		nextToken = strconv.Itoa(int(offset) + len(entries))
	}

	return &ledgerv1.GetWalletHistoryResponse{
		Entries:       protoEntries,
		NextPageToken: nextToken,
	}, nil
}

func classifyError(err error) string {
	switch {
	case errors.Is(err, ledger.ErrInsufficientFunds):
		return "insufficient_funds"
	case errors.Is(err, ledger.ErrCurrencyMismatch):
		return "currency_mismatch"
	case errors.Is(err, ledger.ErrDuplicateTransaction):
		return "duplicate"
	case errors.Is(err, ledger.ErrInvalidAmount),
		errors.Is(err, ledger.ErrSameAccount),
		errors.Is(err, ledger.ErrInvalidCurrencyCode),
		errors.Is(err, ledger.ErrInvalidAccountType):
		return "invalid_input"
	case errors.Is(err, ledger.ErrAccountNotFound),
		errors.Is(err, ledger.ErrTransactionNotFound):
		return "not_found"
	default:
		return "internal"
	}
}
