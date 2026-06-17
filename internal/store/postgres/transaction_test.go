package postgres_test

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
	"github.com/mohammad-farrokhnia/go-ledger/internal/store/postgres"
)

func newKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func setupStore(t *testing.T) *postgres.Store {
	t.Helper()

	dsn := os.Getenv("DB_DSN_TEST")
	if dsn == "" {
		t.Skip("DB_DSN_TEST not set — skipping integration test")
	}

	s, err := postgres.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	t.Cleanup(s.Close)
	return s
}

func seedWallets(t *testing.T, s *postgres.Store, amount int64) (string, string) {
	t.Helper()
	ctx := context.Background()

	sys, err := s.CreateAccount(ctx, ledger.CreateAccountParams{
		Name: "test-system", Type: ledger.AccountTypeSystem, CurrencyCode: "USD",
	})
	if err != nil {
		t.Fatalf("create system account: %v", err)
	}

	usr, err := s.CreateAccount(ctx, ledger.CreateAccountParams{
		Name: "test-user", Type: ledger.AccountTypeUser, CurrencyCode: "USD",
	})
	if err != nil {
		t.Fatalf("create user account: %v", err)
	}

	_, err = s.CreateTransaction(ctx, ledger.CreateTransactionParams{
		IdempotencyKey: usr.ID,
		FromAccountID:  sys.ID,
		ToAccountID:    usr.ID,
		Amount:         amount,
		CurrencyCode:   "USD",
	})
	if err != nil {
		t.Fatalf("seed funding: %v", err)
	}

	return sys.ID, usr.ID
}

func TestCreateTransaction_Success(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	sysID, userID := seedWallets(t, s, 1000)

	tx, err := s.CreateTransaction(ctx, ledger.CreateTransactionParams{
		IdempotencyKey: newKey(),
		FromAccountID:  userID,
		ToAccountID:    sysID,
		Amount:         100,
		CurrencyCode:   "USD",
	})
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}
	if tx.Status != ledger.TransactionStatusCompleted {
		t.Errorf("status: got %q, want COMPLETED", tx.Status)
	}

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if balance != 900 {
		t.Errorf("balance after tx: got %d, want 900", balance)
	}
}

func TestCreateTransaction_InsufficientFunds(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	sysID, userID := seedWallets(t, s, 50)

	_, err := s.CreateTransaction(ctx, ledger.CreateTransactionParams{
		IdempotencyKey: newKey(),
		FromAccountID:  userID,
		ToAccountID:    sysID,
		Amount:         100,
		CurrencyCode:   "USD",
	})
	if !errors.Is(err, ledger.ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got: %v", err)
	}

	balance, _ := s.GetBalance(ctx, userID)
	if balance != 50 {
		t.Errorf("balance after failed tx: got %d, want 50 (unchanged)", balance)
	}
}

func TestCreateTransaction_Idempotency(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	sysID, userID := seedWallets(t, s, 1000)

	params := ledger.CreateTransactionParams{
		IdempotencyKey: newKey(),
		FromAccountID:  userID,
		ToAccountID:    sysID,
		Amount:         100,
		CurrencyCode:   "USD",
	}

	tx1, err := s.CreateTransaction(ctx, params)
	if err != nil {
		t.Fatalf("first CreateTransaction: %v", err)
	}

	tx2, err := s.CreateTransaction(ctx, params)
	if err != nil {
		t.Fatalf("duplicate CreateTransaction: %v", err)
	}

	if tx1.ID != tx2.ID {
		t.Errorf("idempotency broken: got different IDs %s vs %s", tx1.ID, tx2.ID)
	}

	balance, _ := s.GetBalance(ctx, userID)
	if balance != 900 {
		t.Errorf("balance after duplicate tx: got %d, want 900 (deducted once)", balance)
	}
}

func TestCreateTransaction_CurrencyMismatch(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	usd, _ := s.CreateAccount(ctx, ledger.CreateAccountParams{
		Name: "usd-sys", Type: ledger.AccountTypeSystem, CurrencyCode: "USD",
	})
	irr, _ := s.CreateAccount(ctx, ledger.CreateAccountParams{
		Name: "irr-sys", Type: ledger.AccountTypeSystem, CurrencyCode: "IRR",
	})

	_, err := s.CreateTransaction(ctx, ledger.CreateTransactionParams{
		IdempotencyKey: newKey(),
		FromAccountID:  usd.ID,
		ToAccountID:    irr.ID,
		Amount:         100,
		CurrencyCode:   "USD",
	})
	if !errors.Is(err, ledger.ErrCurrencyMismatch) {
		t.Errorf("expected ErrCurrencyMismatch, got: %v", err)
	}
}

func TestCreateTransaction_ConcurrentDeductions(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	const (
		numGoroutines = 50
		deductAmount  = 10
		initialFunds  = numGoroutines * deductAmount
	)

	sysID, userID := seedWallets(t, s, initialFunds)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var succeeded, failed int

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(i int) {
			defer wg.Done()

			key := newKey()

			_, err := s.CreateTransaction(ctx, ledger.CreateTransactionParams{
				IdempotencyKey: key,
				FromAccountID:  userID,
				ToAccountID:    sysID,
				Amount:         deductAmount,
				CurrencyCode:   "USD",
			})

			mu.Lock()
			if err == nil {
				succeeded++
			} else {
				failed++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	finalBalance, _ := s.GetBalance(ctx, userID)

	t.Logf("goroutines=%d succeeded=%d failed=%d finalBalance=%d",
		numGoroutines, succeeded, failed, finalBalance)

	if succeeded != numGoroutines {
		t.Errorf("succeeded: got %d, want %d", succeeded, numGoroutines)
	}
	if failed != 0 {
		t.Errorf("failed: got %d, want 0", failed)
	}

	if finalBalance != 0 {
		t.Errorf("CORRECTNESS FAILURE: final balance=%d want 0\n"+
			"This means SELECT FOR UPDATE is not preventing a race condition.", finalBalance)
	}
}

func TestCreateTransaction_ConcurrentContention(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	const (
		numGoroutines = 100
		deductAmount  = 10
		initialFunds  = 500
	)

	sysID, userID := seedWallets(t, s, initialFunds)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var succeeded, failed int

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(i int) {
			defer wg.Done()

			key := newKey()

			_, err := s.CreateTransaction(ctx, ledger.CreateTransactionParams{
				IdempotencyKey: key,
				FromAccountID:  userID,
				ToAccountID:    sysID,
				Amount:         deductAmount,
				CurrencyCode:   "USD",
			})

			mu.Lock()
			if err == nil {
				succeeded++
			} else {
				failed++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	finalBalance, _ := s.GetBalance(ctx, userID)

	t.Logf("goroutines=%d succeeded=%d failed=%d finalBalance=%d",
		numGoroutines, succeeded, failed, finalBalance)

	expectedSuccess := initialFunds / deductAmount
	if succeeded != expectedSuccess {
		t.Errorf("succeeded: got %d, want %d", succeeded, expectedSuccess)
	}
	if finalBalance != 0 {
		t.Errorf("CORRECTNESS FAILURE: final balance=%d want 0", finalBalance)
	}
}
