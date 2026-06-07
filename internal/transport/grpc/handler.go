package grpc

import (
	"context"

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
		return nil, domainErrorToGRPC(err)
	}

	return &ledgerv1.CreateWalletResponse{Wallet: accountToProto(acc)}, nil
}

func (h *Handler) GetBalance(ctx context.Context, req *ledgerv1.GetBalanceRequest) (*ledgerv1.GetBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	balance, err := h.svc.GetBalance(ctx, req.WalletId)
	if err != nil {
		return nil, domainErrorToGRPC(err)
	}

	acc, err := h.svc.GetAccount(ctx, req.WalletId)
	if err != nil {
		return nil, domainErrorToGRPC(err)
	}

	return &ledgerv1.GetBalanceResponse{
		WalletId:     req.WalletId,
		Balance:      balance,
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
		return nil, domainErrorToGRPC(err)
	}

	return &ledgerv1.CreateTransactionResponse{Transaction: transactionToProto(tx)}, nil
}

func (h *Handler) GetWalletHistory(ctx context.Context, req *ledgerv1.GetWalletHistoryRequest) (*ledgerv1.GetWalletHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	entries, err := h.svc.GetWalletHistory(ctx, req.WalletId, req.PageSize, 0)
	if err != nil {
		return nil, domainErrorToGRPC(err)
	}

	protoEntries := make([]*ledgerv1.Entry, len(entries))
	for i, e := range entries {
		protoEntries[i] = entryToProto(e)
	}

	return &ledgerv1.GetWalletHistoryResponse{Entries: protoEntries}, nil
}
