package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mohammad-farrokhnia/ledger/internal/i18n"
	"github.com/mohammad-farrokhnia/ledger/internal/ledger"
)

var grpcToHTTP = map[codes.Code]int{
	codes.OK:                 http.StatusOK,
	codes.NotFound:           http.StatusNotFound,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.FailedPrecondition: http.StatusUnprocessableEntity,
	codes.AlreadyExists:      http.StatusConflict,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.Unavailable:        http.StatusServiceUnavailable,
}

var grpcToMessageCode = map[string]i18n.MessageCode{
	ledger.ErrAccountNotFound.Error():      i18n.MsgAccountNotFound,
	ledger.ErrTransactionNotFound.Error():  i18n.MsgTransactionNotFound,
	ledger.ErrInsufficientFunds.Error():    i18n.MsgInsufficientFunds,
	ledger.ErrCurrencyMismatch.Error():     i18n.MsgCurrencyMismatch,
	ledger.ErrDuplicateTransaction.Error(): i18n.MsgDuplicateTransaction,
	ledger.ErrInvalidAmount.Error():        i18n.MsgInvalidAmount,
	ledger.ErrSameAccount.Error():          i18n.MsgSameAccount,
	ledger.ErrInvalidAccountType.Error():   i18n.MsgInvalidAccountType,
	ledger.ErrInvalidCurrencyCode.Error():  i18n.MsgInvalidCurrencyCode,
}

type Meta struct {
	AppName     string `json:"appName"`
	Version     string `json:"version"`
	Timestamp   string `json:"timestamp"`
	MessageCode string `json:"messageCode"`
	Message     string `json:"message"`
}

func newMeta(acceptLang string, code i18n.MessageCode) Meta {
	return Meta{
		AppName:     appName,
		Version:     appVersion,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		MessageCode: string(code),
		Message:     i18n.Translate(acceptLang, code),
	}
}

const (
	appName    = "ledger"
	appVersion = "1.0.0"
)

func CustomErrorHandler(
	_ context.Context,
	_ *runtime.ServeMux,
	_ runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	lang := r.Header.Get("Accept-Language")
	s, _ := status.FromError(err)

	httpCode, ok := grpcToHTTP[s.Code()]
	if !ok {
		httpCode = http.StatusInternalServerError
	}

	msgCode, errCode := resolveErrorCodes(s.Code(), s.Message())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"error": map[string]any{"code": errCode},
		"meta":  newMeta(lang, msgCode),
	})
}
func resolveErrorCodes(code codes.Code, message string) (i18n.MessageCode, string) {
	if msgCode, ok := grpcToMessageCode[message]; ok {
		return msgCode, string(msgCode)
	}

	switch code {
	case codes.InvalidArgument:
		return i18n.MsgValidationFailed, string(i18n.MsgValidationFailed)
	case codes.NotFound:
		return i18n.MsgNotFound, string(i18n.MsgNotFound)
	case codes.AlreadyExists:
		return i18n.MsgConflict, string(i18n.MsgConflict)
	case codes.FailedPrecondition:
		return i18n.MsgInsufficientFunds, string(i18n.MsgInsufficientFunds)
	default:
		return i18n.MsgInternalError, string(i18n.MsgInternalError)
	}
}

func healthHandler(ping PingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		dbOK := ping(r.Context()) == nil

		w.Header().Set("Content-Type", "application/json")

		if !dbOK {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"data": map[string]any{
					"status": "unavailable",
					"checks": map[string]string{"postgres": "unavailable"},
				},
				"meta": newMeta(lang, i18n.MsgInternalError),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"data": map[string]any{
				"status":  "ok",
				"version": appVersion,
				"app":     appName,
				"checks":  map[string]string{"postgres": "ok"},
			},
			"meta": newMeta(lang, i18n.MsgHealthOK),
		})
	}
}
