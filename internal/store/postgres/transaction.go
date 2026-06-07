package postgres

import (
	"context"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"github.com/mohammad-farrokhnia/go-ledger/internal/store"
)

func (s *Store) CreateTransaction(ctx context.Context, params store.CreateTransactionParams) (ledger.Transaction, error) {
	panic("not implemented — Task 3.3")
}

func (s *Store) GetTransaction(ctx context.Context, id string) (ledger.Transaction, error) {
	panic("not implemented — Task 3.3")
}

func (s *Store) GetTransactionByIdempotencyKey(ctx context.Context, key string) (ledger.Transaction, error) {
	panic("not implemented — Task 3.3")
}

func (s *Store) GetWalletHistory(ctx context.Context, params store.GetWalletHistoryParams) ([]ledger.Entry, error) {
	panic("not implemented — Task 3.4")
}
