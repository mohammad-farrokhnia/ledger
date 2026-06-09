package ledger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type mockStore struct {
	accounts     map[string]ledger.Account
	transactions map[string]ledger.Transaction
	createErr    error
	getErr       error
}

func newMockStore() *mockStore {
	return &mockStore{
		accounts:     make(map[string]ledger.Account),
		transactions: make(map[string]ledger.Transaction),
	}
}

func (m *mockStore) CreateAccount(_ context.Context, p ledger.CreateAccountParams) (ledger.Account, error) {
	if m.createErr != nil {
		return ledger.Account{}, m.createErr
	}
	acc := ledger.Account{
		ID:           "acc-" + p.Name,
		Name:         p.Name,
		Type:         p.Type,
		CurrencyCode: p.CurrencyCode,
		Balance:      0,
	}
	m.accounts[acc.ID] = acc
	return acc, nil
}

func (m *mockStore) GetAccount(_ context.Context, id string) (ledger.Account, error) {
	if m.getErr != nil {
		return ledger.Account{}, m.getErr
	}
	acc, ok := m.accounts[id]
	if !ok {
		return ledger.Account{}, ledger.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockStore) GetBalance(_ context.Context, id string) (int64, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return 0, ledger.ErrAccountNotFound
	}
	return acc.Balance, nil
}

func (m *mockStore) CreateTransaction(_ context.Context, p ledger.CreateTransactionParams) (ledger.Transaction, error) {
	if m.createErr != nil {
		return ledger.Transaction{}, m.createErr
	}
	tx := ledger.Transaction{
		ID:             "tx-" + p.IdempotencyKey,
		IdempotencyKey: p.IdempotencyKey,
		FromAccountID:  p.FromAccountID,
		ToAccountID:    p.ToAccountID,
		Amount:         p.Amount,
		CurrencyCode:   p.CurrencyCode,
		Status:         ledger.TransactionStatusCompleted,
	}
	m.transactions[p.IdempotencyKey] = tx
	return tx, nil
}

func (m *mockStore) GetTransaction(_ context.Context, id string) (ledger.Transaction, error) {
	for _, tx := range m.transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return ledger.Transaction{}, ledger.ErrTransactionNotFound
}

func (m *mockStore) GetTransactionByIdempotencyKey(_ context.Context, key string) (ledger.Transaction, error) {
	tx, ok := m.transactions[key]
	if !ok {
		return ledger.Transaction{}, ledger.ErrTransactionNotFound
	}
	return tx, nil
}

func (m *mockStore) GetWalletHistory(_ context.Context, _ ledger.GetWalletHistoryParams) ([]ledger.Entry, error) {
	return nil, nil
}


func TestService_CreateAccount_EmptyName(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateAccount(context.Background(), "", ledger.AccountTypeUser, "USD")
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	var valErr *ledger.ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestService_CreateAccount_InvalidCurrency(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	cases := []string{"", "US", "USDD", "123", "u$d"}
	for _, code := range cases {
		_, err := svc.CreateAccount(context.Background(), "test", ledger.AccountTypeUser, code)
		if err == nil {
			t.Errorf("expected error for currency %q, got nil", code)
		}
	}
}

func TestService_CreateAccount_InvalidType(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateAccount(context.Background(), "test", "UNKNOWN_TYPE", "USD")
	if !errors.Is(err, ledger.ErrInvalidAccountType) {
		t.Errorf("expected ErrInvalidAccountType, got %v", err)
	}
}

func TestService_CreateTransaction_ZeroAmount(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateTransaction(context.Background(), ledger.CreateTransactionInput{
		IdempotencyKey: "key-1",
		FromAccountID:  "a",
		ToAccountID:    "b",
		Amount:         0,
		CurrencyCode:   "USD",
	})
	if !errors.Is(err, ledger.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestService_CreateTransaction_NegativeAmount(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateTransaction(context.Background(), ledger.CreateTransactionInput{
		IdempotencyKey: "key-2",
		FromAccountID:  "a",
		ToAccountID:    "b",
		Amount:         -100,
		CurrencyCode:   "USD",
	})
	if !errors.Is(err, ledger.ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestService_CreateTransaction_SameAccount(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateTransaction(context.Background(), ledger.CreateTransactionInput{
		IdempotencyKey: "key-3",
		FromAccountID:  "same",
		ToAccountID:    "same",
		Amount:         100,
		CurrencyCode:   "USD",
	})
	if !errors.Is(err, ledger.ErrSameAccount) {
		t.Errorf("expected ErrSameAccount, got %v", err)
	}
}

func TestService_CreateTransaction_EmptyIdempotencyKey(t *testing.T) {
	svc := ledger.NewService(newMockStore())

	_, err := svc.CreateTransaction(context.Background(), ledger.CreateTransactionInput{
		IdempotencyKey: "",
		FromAccountID:  "a",
		ToAccountID:    "b",
		Amount:         100,
		CurrencyCode:   "USD",
	})

	var valErr *ledger.ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %v", err)
	}
}

func TestService_GetWalletHistory_DefaultLimit(t *testing.T) {
	store := newMockStore()
	svc := ledger.NewService(store)

	store.accounts["test-id"] = ledger.Account{ID: "test-id", CurrencyCode: "USD"}

	_, err := svc.GetWalletHistory(context.Background(), "test-id", 0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_GetWalletHistory_CapLimit(t *testing.T) {
	store := newMockStore()
	svc := ledger.NewService(store)

	store.accounts["test-id"] = ledger.Account{ID: "test-id"}

	_, err := svc.GetWalletHistory(context.Background(), "test-id", 200, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
