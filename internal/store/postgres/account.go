package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/mohammad-farrokhnia/ledger/internal/ledger"
)

func (s *Store) CreateAccount(ctx context.Context, params ledger.CreateAccountParams) (ledger.Account, error) {
	const q = `
		INSERT INTO accounts (name, type, currency_code)
		VALUES ($1, $2, $3)
		RETURNING id, name, type, currency_code, balance, created_at, updated_at
	`

	var acc ledger.Account
	var accountType string

	row := s.pool.QueryRow(ctx, q, params.Name, string(params.Type), params.CurrencyCode)
	err := row.Scan(
		&acc.ID,
		&acc.Name,
		&accountType,
		&acc.CurrencyCode,
		&acc.Balance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		return ledger.Account{}, fmt.Errorf("postgres: create account: %w", err)
	}

	acc.Type = ledger.AccountType(accountType)
	return acc, nil
}

func (s *Store) GetAccount(ctx context.Context, id string) (ledger.Account, error) {
	const q = `
		SELECT id, name, type, currency_code, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var acc ledger.Account
	var accountType string

	row := s.pool.QueryRow(ctx, q, id)
	err := row.Scan(
		&acc.ID,
		&acc.Name,
		&accountType,
		&acc.CurrencyCode,
		&acc.Balance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.Account{}, ledger.ErrAccountNotFound
		}
		return ledger.Account{}, fmt.Errorf("postgres: get account: %w", err)
	}

	acc.Type = ledger.AccountType(accountType)
	return acc, nil
}

func (s *Store) GetBalance(ctx context.Context, accountID string) (int64, error) {
	const q = `SELECT balance FROM accounts WHERE id = $1`

	var balance int64
	err := s.pool.QueryRow(ctx, q, accountID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ledger.ErrAccountNotFound
		}
		return 0, fmt.Errorf("postgres: get balance: %w", err)
	}

	return balance, nil
}
