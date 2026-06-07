package postgres

import (
	"context"
	"fmt"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"github.com/mohammad-farrokhnia/go-ledger/internal/store"
)

func (s *Store) GetWalletHistory(ctx context.Context, params store.GetWalletHistoryParams) ([]ledger.Entry, error) {
	const q = `
		SELECT id, account_id, transaction_id, amount, created_at
		FROM entries
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, q, params.AccountID, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("postgres: get wallet history: %w", err)
	}
	defer rows.Close()

	var entries []ledger.Entry
	for rows.Next() {
		var e ledger.Entry
		if err = rows.Scan(&e.ID, &e.AccountID, &e.TransactionID, &e.Amount, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan entry: %w", err)
		}
		entries = append(entries, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: wallet history rows: %w", err)
	}

	return entries, nil
}
