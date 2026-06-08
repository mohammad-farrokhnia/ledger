package grpc

import (
	"context"
	"log/slog"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ledgerv1 "github.com/mohammad-farrokhnia/go-ledger/api/proto/ledger/v1"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
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
		slog.ErrorContext(ctx, "CreateWallet failed", "error", err)
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
		slog.ErrorContext(ctx, "GetBalance: get account failed", "error", err, "wallet_id", req.WalletId)
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

	tx, err := h.svc.CreateTransaction(ctx, ledger.CreateTransactionInput{
		IdempotencyKey: req.IdempotencyKey,
		FromAccountID:  req.FromAccountId,
		ToAccountID:    req.ToAccountId,
		Amount:         req.Amount,
		CurrencyCode:   req.CurrencyCode,
	})
	if err != nil {
		slog.ErrorContext(ctx, "CreateTransaction failed",
			"error", err,
			"idempotency_key", req.IdempotencyKey,
		)
		return nil, domainErrorToGRPC(err)
	}

	slog.InfoContext(ctx, "transaction created",
		"transaction_id", tx.ID,
		"from", tx.FromAccountID,
		"to", tx.ToAccountID,
		"amount", tx.Amount,
		"currency", tx.CurrencyCode,
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