package ledger

import "errors"

var (
	ErrAccountNotFound = errors.New("account not found")

	ErrInsufficientFunds = errors.New("insufficient funds")

	ErrCurrencyMismatch = errors.New("currency mismatch between accounts")

	ErrDuplicateTransaction = errors.New("transaction with this idempotency key already exists")

	ErrInvalidAmount = errors.New("amount must be greater than zero")

	ErrSameAccount = errors.New("source and destination accounts must be different")

	ErrInvalidAccountType = errors.New("invalid account type")

	ErrInvalidCurrencyCode = errors.New("currency code must be a 3-letter ISO 4217 code")
)
