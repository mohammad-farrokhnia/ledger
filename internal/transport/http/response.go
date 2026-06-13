package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mohammad-farrokhnia/go-ledger/internal/i18n"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

const (
	appName    = "go-ledger"
	appVersion = "1.0.0"
)

type Meta struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type Response struct {
	Data map[string]any `json:"data"`
	Meta Meta           `json:"meta"`
}

func newMeta(httpCode int, message string) Meta {
	return Meta{
		Name:      appName,
		Version:   appVersion,
		Code:      httpCode,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func SuccessResponse(httpCode int, data map[string]any) Response {
	d := map[string]any{"success": true}
	for k, v := range data {
		d[k] = v
	}
	return Response{
		Data: d,
		Meta: newMeta(httpCode, http.StatusText(httpCode)),
	}
}

func errorResponse(httpCode int, message string) Response {
	return Response{
		Data: map[string]any{"success": false},
		Meta: newMeta(httpCode, message),
	}
}

var grpcToHTTPStatus = map[codes.Code]int{
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

func CustomErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	m runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	lang := i18n.Parse(r.Header.Get("Accept-Language"))
	s, _ := status.FromError(err)
	code := s.Code()

	httpStatus, ok := grpcToHTTPStatus[code]
	if !ok {
		httpStatus = http.StatusInternalServerError
	}

	message := localizedMessage(s.Message(), lang)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	jsonEncode(w, errorResponse(httpStatus, message))
}

func localizedMessage(grpcMessage string, lang i18n.Lang) string {
	domainErrors := []error{
		ledger.ErrAccountNotFound,
		ledger.ErrTransactionNotFound,
		ledger.ErrInsufficientFunds,
		ledger.ErrCurrencyMismatch,
		ledger.ErrDuplicateTransaction,
		ledger.ErrInvalidAmount,
		ledger.ErrSameAccount,
		ledger.ErrInvalidAccountType,
		ledger.ErrInvalidCurrencyCode,
	}

	for _, domainErr := range domainErrors {
		if grpcMessage == domainErr.Error() {
			return i18n.Message(domainErr, lang)
		}
	}

	return i18n.Message(errors.New("internal"), lang)
}
