package store

import (
	"context"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type Store interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (ledger.Account, error)

	GetAccount(ctx context.Context, id string) (ledger.Account, error)

	GetBalance(ctx context.Context, accountID string) (int64, error)

	CreateTransaction(ctx context.Context, params CreateTransactionParams) (ledger.Transaction, error)

	GetTransaction(ctx context.Context, id string) (ledger.Transaction, error)

	GetTransactionByIdempotencyKey(ctx context.Context, key string) (ledger.Transaction, error)

	GetWalletHistory(ctx context.Context, params GetWalletHistoryParams) ([]ledger.Entry, error)
}

type CreateAccountParams struct {
	Name         string
	Type         ledger.AccountType
	CurrencyCode string
}

type CreateTransactionParams struct {
	IdempotencyKey string
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	CurrencyCode   string
}

type GetWalletHistoryParams struct {
	AccountID string
	// Default: 20, max: 100.
	Limit int32
	// Default: 0.
	Offset int32
}
