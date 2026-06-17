package ledger

import "time"

type AccountType string

const (
	AccountTypeUser     AccountType = "USER"
	AccountTypeSystem   AccountType = "SYSTEM"
	AccountTypeProvider AccountType = "PROVIDER"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

type Account struct {
	ID           string
	Name         string
	Type         AccountType
	CurrencyCode string
	Balance      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Transaction struct {
	ID             string
	IdempotencyKey string
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	CurrencyCode   string
	Status         TransactionStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Entry struct {
	ID            string
	AccountID     string
	TransactionID string
	// Positive = CREDIT (money in). Negative = DEBIT (money out).
	Amount    int64
	CreatedAt time.Time
}
