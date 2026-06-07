package ledger

import "context"

type Store interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (Account, error)
	GetAccount(ctx context.Context, id string) (Account, error)
	GetBalance(ctx context.Context, accountID string) (int64, error)
	CreateTransaction(ctx context.Context, params CreateTransactionParams) (Transaction, error)
	GetTransaction(ctx context.Context, id string) (Transaction, error)
	GetTransactionByIdempotencyKey(ctx context.Context, key string) (Transaction, error)
	GetWalletHistory(ctx context.Context, params GetWalletHistoryParams) ([]Entry, error)
}

type (
	CreateAccountParams struct {
		Name         string
		Type         AccountType
		CurrencyCode string
	}

	CreateTransactionParams struct {
		IdempotencyKey string
		FromAccountID  string
		ToAccountID    string
		Amount         int64
		CurrencyCode   string
	}

	GetWalletHistoryParams struct {
		AccountID string
		Limit     int32
		Offset    int32
	}
)
