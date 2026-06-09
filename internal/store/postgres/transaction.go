package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type accountSnapshot struct {
	id           string
	accountType  ledger.AccountType
	currencyCode string
	balance      int64
}

func (s *Store) CreateTransaction(ctx context.Context, params ledger.CreateTransactionParams) (ledger.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer rollBack(ctx, tx)

	ids := []string{params.FromAccountID, params.ToAccountID}
	sort.Strings(ids)

	const lockQ = `
		SELECT id, type, currency_code, balance
		FROM accounts
		WHERE id = ANY($1::uuid[])
		ORDER BY id
		FOR UPDATE
	`

	rows, err := tx.Query(ctx, lockQ, ids)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: lock accounts: %w", err)
	}

	var snapshots []accountSnapshot
	for rows.Next() {
		var snap accountSnapshot
		var accType string
		if err = rows.Scan(&snap.id, &accType, &snap.currencyCode, &snap.balance); err != nil {
			rows.Close()
			return ledger.Transaction{}, fmt.Errorf("postgres: scan account: %w", err)
		}
		snap.accountType = ledger.AccountType(accType)
		snapshots = append(snapshots, snap)
	}
	rows.Close()

	if err = rows.Err(); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: lock accounts rows: %w", err)
	}

	if len(snapshots) != 2 {
		return ledger.Transaction{}, ledger.ErrTransactionNotFound
	}

	var from, to accountSnapshot
	if snapshots[0].id == params.FromAccountID {
		from, to = snapshots[0], snapshots[1]
	} else {
		from, to = snapshots[1], snapshots[0]
	}

	if from.currencyCode != to.currencyCode {
		return ledger.Transaction{}, ledger.ErrCurrencyMismatch
	}

	if from.currencyCode != params.CurrencyCode {
		return ledger.Transaction{}, ledger.ErrCurrencyMismatch
	}

	if from.accountType != ledger.AccountTypeSystem {
		if from.balance < params.Amount {
			return ledger.Transaction{}, ledger.ErrInsufficientFunds
		}
	}

	const insertTxQ = `
		INSERT INTO transactions
			(idempotency_key, from_account_id, to_account_id, amount, currency_code, status)
		VALUES
			($1, $2, $3, $4, $5, 'PENDING')
		ON CONFLICT (idempotency_key) DO NOTHING
	`

	tag, err := tx.Exec(ctx, insertTxQ,
		params.IdempotencyKey,
		params.FromAccountID,
		params.ToAccountID,
		params.Amount,
		params.CurrencyCode,
	)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: insert transaction: %w", err)
	}

	if tag.RowsAffected() == 0 {
		rollBack(ctx, tx)
		return s.GetTransactionByIdempotencyKey(ctx, params.IdempotencyKey)
	}

	var txID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM transactions WHERE idempotency_key = $1`,
		params.IdempotencyKey,
	).Scan(&txID)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: fetch transaction id: %w", err)
	}

	const insertEntryQ = `
		INSERT INTO entries (account_id, transaction_id, amount)
		VALUES ($1, $2, $3)
	`

	if _, err = tx.Exec(ctx, insertEntryQ, params.FromAccountID, txID, -params.Amount); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: insert debit entry: %w", err)
	}

	if _, err = tx.Exec(ctx, insertEntryQ, params.ToAccountID, txID, params.Amount); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: insert credit entry: %w", err)
	}

	const updateBalanceQ = `UPDATE accounts SET balance = balance + $1 WHERE id = $2`

	if _, err = tx.Exec(ctx, updateBalanceQ, -params.Amount, params.FromAccountID); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: deduct balance: %w", err)
	}

	if _, err = tx.Exec(ctx, updateBalanceQ, params.Amount, params.ToAccountID); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: add balance: %w", err)
	}

	auditPayload, err := json.Marshal(map[string]any{
		"transaction_id":  txID,
		"from_account_id": params.FromAccountID,
		"to_account_id":   params.ToAccountID,
		"amount":          params.Amount,
		"currency_code":   params.CurrencyCode,
	})
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: marshal audit payload: %w", err)
	}

	const insertOutboxQ = `
		INSERT INTO audit_outbox (action_type, payload)
		VALUES ($1, $2)
	`
	if _, err = tx.Exec(ctx, insertOutboxQ, "transaction.created", auditPayload); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: insert audit outbox: %w", err)
	}
	
	const completeTxQ = `
		UPDATE transactions
		SET status = 'COMPLETED'
		WHERE id = $1
		RETURNING id, idempotency_key, from_account_id, to_account_id,
		          amount, currency_code, status, created_at, updated_at
	`

	result, err := scanTransaction(tx.QueryRow(ctx, completeTxQ, txID))
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: complete transaction: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return ledger.Transaction{}, fmt.Errorf("postgres: commit: %w", err)
	}

	return result, nil
}

func (s *Store) GetTransaction(ctx context.Context, id string) (ledger.Transaction, error) {
	const q = `
		SELECT id, idempotency_key, from_account_id, to_account_id,
		       amount, currency_code, status, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	result, err := scanTransaction(s.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.Transaction{}, ledger.ErrTransactionNotFound
		}
		return ledger.Transaction{}, fmt.Errorf("postgres: get transaction: %w", err)
	}

	return result, nil
}

func (s *Store) GetTransactionByIdempotencyKey(ctx context.Context, key string) (ledger.Transaction, error) {
	const q = `
		SELECT id, idempotency_key, from_account_id, to_account_id,
		       amount, currency_code, status, created_at, updated_at
		FROM transactions
		WHERE idempotency_key = $1
	`

	result, err := scanTransaction(s.pool.QueryRow(ctx, q, key))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.Transaction{}, ledger.ErrDuplicateTransaction
		}
		return ledger.Transaction{}, fmt.Errorf("postgres: get transaction by idempotency key: %w", err)
	}

	return result, nil
}

func scanTransaction(row interface {
	Scan(dest ...any) error
}) (ledger.Transaction, error) {
	var t ledger.Transaction
	var status string

	err := row.Scan(
		&t.ID,
		&t.IdempotencyKey,
		&t.FromAccountID,
		&t.ToAccountID,
		&t.Amount,
		&t.CurrencyCode,
		&status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return ledger.Transaction{}, err
	}

	t.Status = ledger.TransactionStatus(status)
	return t, nil
}

func rollBack(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.WarnContext(ctx, "postgres: unexpected rollback error", "error", err)
	}
}
