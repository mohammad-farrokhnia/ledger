// Package apperr provides the AppError type for go-ledger.
// Mirrors LibreCore's pkg/errors: typed error codes, error types,
// HTTP status mapping, and context for additional details.
package apperr

import (
	"fmt"
	"net/http"

	"github.com/mohammad-farrokhnia/go-ledger/internal/i18n"
)

// Type classifies the error for logging and HTTP status mapping.
type Type string

const (
	TypeNotFound     Type = "not_found"
	TypeValidation   Type = "validation"
	TypeConflict     Type = "conflict"
	TypeInternal     Type = "internal"
	TypeUnauthorized Type = "unauthorized"
	TypeForbidden    Type = "forbidden"
)

// Code is a stable string identifier for each error condition.
// Clients can rely on these — they never change.
const (
	ErrAccountNotFound     = "ACCOUNT_NOT_FOUND"
	ErrTransactionNotFound = "TRANSACTION_NOT_FOUND"
	ErrInsufficientFunds   = "INSUFFICIENT_FUNDS"
	ErrCurrencyMismatch    = "CURRENCY_MISMATCH"
	ErrDuplicateTransaction = "DUPLICATE_TRANSACTION"
	ErrInvalidAmount       = "INVALID_AMOUNT"
	ErrSameAccount         = "SAME_ACCOUNT"
	ErrInvalidAccountType  = "INVALID_ACCOUNT_TYPE"
	ErrInvalidCurrencyCode = "INVALID_CURRENCY_CODE"
	ErrInternal            = "INTERNAL_ERROR"
	ErrBadRequest          = "BAD_REQUEST"
)

// codeToMessageCode maps error codes to i18n message codes.
var codeToMessageCode = map[string]i18n.MessageCode{
	ErrAccountNotFound:      i18n.MsgAccountNotFound,
	ErrTransactionNotFound:  i18n.MsgTransactionNotFound,
	ErrInsufficientFunds:    i18n.MsgInsufficientFunds,
	ErrCurrencyMismatch:     i18n.MsgCurrencyMismatch,
	ErrDuplicateTransaction: i18n.MsgDuplicateTransaction,
	ErrInvalidAmount:        i18n.MsgInvalidAmount,
	ErrSameAccount:          i18n.MsgSameAccount,
	ErrInvalidAccountType:   i18n.MsgInvalidAccountType,
	ErrInvalidCurrencyCode:  i18n.MsgInvalidCurrencyCode,
	ErrInternal:             i18n.MsgInternalError,
	ErrBadRequest:           i18n.MsgBadRequest,
}

// MessageCode returns the i18n code for this error code.
func MessageCode(errCode string) i18n.MessageCode {
	if code, ok := codeToMessageCode[errCode]; ok {
		return code
	}
	return i18n.MsgInternalError
}

// AppError is the single error type used across all layers.
type AppError struct {
	code       string
	errType    Type
	underlying error
	ctx        map[string]any
}

func New(code string, t Type) *AppError {
	return &AppError{code: code, errType: t, ctx: make(map[string]any)}
}

func NewWithErr(code string, t Type, underlying error) *AppError {
	return &AppError{code: code, errType: t, underlying: underlying, ctx: make(map[string]any)}
}

func (e *AppError) Error() string {
	if e.underlying != nil {
		return fmt.Sprintf("%s: %v", e.code, e.underlying)
	}
	return e.code
}

func (e *AppError) Code() string    { return e.code }
func (e *AppError) Type() Type      { return e.errType }
func (e *AppError) Unwrap() error   { return e.underlying }

func (e *AppError) WithContext(key string, val any) *AppError {
	e.ctx[key] = val
	return e
}

func (e *AppError) Context() map[string]any { return e.ctx }

func (e *AppError) HTTPStatus() int {
	switch e.errType {
	case TypeNotFound:
		return http.StatusNotFound
	case TypeValidation:
		return http.StatusBadRequest
	case TypeConflict:
		return http.StatusConflict
	case TypeUnauthorized:
		return http.StatusUnauthorized
	case TypeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// Constructors.
func NotFound(code string) *AppError     { return New(code, TypeNotFound) }
func Validation(code string) *AppError   { return New(code, TypeValidation) }
func Conflict(code string) *AppError     { return New(code, TypeConflict) }
func Internal(code string) *AppError     { return New(code, TypeInternal) }
func InternalErr(code string, err error) *AppError { return NewWithErr(code, TypeInternal, err) }
