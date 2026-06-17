package ledger

import (
	"context"
	"regexp"
	"strings"
)

var uuidRegex = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
)

type (
	Servicer interface {
		CreateAccount(ctx context.Context, name string, accountType AccountType, currencyCode string) (Account, error)
		GetAccount(ctx context.Context, id string) (Account, error)
		GetBalance(ctx context.Context, accountID string) (int64, error)
		CreateTransaction(ctx context.Context, params CreateTransactionInput) (Transaction, error)
		GetWalletHistory(ctx context.Context, accountID string, limit, offset int32) ([]Entry, error)
	}
	CreateTransactionInput struct {
		IdempotencyKey string
		FromAccountID  string
		ToAccountID    string
		Amount         int64
		CurrencyCode   string
	}
	Service struct {
		store Store
	}
)

func NewService(s Store) *Service {
	return &Service{store: s}
}

func (s *Service) CreateAccount(ctx context.Context, name string, accountType AccountType, currencyCode string) (Account, error) {
	if strings.TrimSpace(name) == "" {
		return Account{}, NewValidationError("account name cannot be empty")
	}

	if err := validateAccountType(accountType); err != nil {
		return Account{}, err
	}

	if err := validateCurrencyCode(currencyCode); err != nil {
		return Account{}, err
	}

	return s.store.CreateAccount(ctx, CreateAccountParams{
		Name:         strings.TrimSpace(name),
		Type:         accountType,
		CurrencyCode: strings.ToUpper(currencyCode),
	})
}

func (s *Service) GetAccount(ctx context.Context, id string) (Account, error) {
	if id == "" {
		return Account{}, NewValidationError("account ID cannot be empty")
	}
	if err := validateUUID(id, "account ID"); err != nil {
		return Account{}, err
	}

	return s.store.GetAccount(ctx, id)
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (int64, error) {
	if accountID == "" {
		return 0, NewValidationError("account ID cannot be empty")
	}
	if err := validateUUID(accountID, "account ID"); err != nil {
		return 0, err
	}

	return s.store.GetBalance(ctx, accountID)
}

func (s *Service) CreateTransaction(ctx context.Context, params CreateTransactionInput) (Transaction, error) {
	if params.Amount <= 0 {
		return Transaction{}, ErrInvalidAmount
	}

	if params.FromAccountID == params.ToAccountID {
		return Transaction{}, ErrSameAccount
	}

	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return Transaction{}, NewValidationError("idempotency key cannot be empty")
	}

	if err := validateCurrencyCode(params.CurrencyCode); err != nil {
		return Transaction{}, err
	}

	if err := validateUUID(params.FromAccountID, "from account ID"); err != nil {
		return Transaction{}, err
	}
	if err := validateUUID(params.ToAccountID, "to account ID"); err != nil {
		return Transaction{}, err
	}

	tx, err := s.store.CreateTransaction(ctx, CreateTransactionParams{
		IdempotencyKey: params.IdempotencyKey,
		FromAccountID:  params.FromAccountID,
		ToAccountID:    params.ToAccountID,
		Amount:         params.Amount,
		CurrencyCode:   strings.ToUpper(params.CurrencyCode),
	})
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func (s *Service) GetWalletHistory(ctx context.Context, accountID string, limit, offset int32) ([]Entry, error) {
	if accountID == "" {
		return nil, NewValidationError("account ID cannot be empty")
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.store.GetWalletHistory(ctx, GetWalletHistoryParams{
		AccountID: accountID,
		Limit:     limit,
		Offset:    offset,
	})
}

func validateCurrencyCode(code string) error {
	upper := strings.ToUpper(code)
	if len(upper) != 3 {
		return ErrInvalidCurrencyCode
	}

	for _, c := range upper {
		if c < 'A' || c > 'Z' {
			return ErrInvalidCurrencyCode
		}
	}

	return nil
}

func validateAccountType(t AccountType) error {
	switch t {
	case AccountTypeUser, AccountTypeSystem, AccountTypeProvider:
		return nil
	default:
		return ErrInvalidAccountType
	}
}

func validateUUID(id, field string) error {
	if id == "" {
		return NewValidationError(field + " cannot be empty")
	}
	if !uuidRegex.MatchString(strings.ToLower(id)) {
		return NewValidationError(field + " must be a valid UUID")
	}
	return nil
}

var _ Servicer = (*Service)(nil)
