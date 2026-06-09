package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mohammad-farrokhnia/go-ledger/internal/i18n"
	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
	Meta    Meta        `json:"meta"`
}

type ErrorDetail struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status"`
}

type Meta struct {
	Timestamp string `json:"timestamp"`
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

func CustomErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	lang := i18n.Parse(r.Header.Get("Accept-Language"))

	s, _ := status.FromError(err)
	code := s.Code()
	httpStatus := grpcToHTTPStatus[code]
	if httpStatus == 0 {
		httpStatus = http.StatusInternalServerError
	}

	message := localizedMessage(s.Message(), lang)

	resp := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:       code.String(),
			Message:    message,
			HTTPStatus: httpStatus,
		},
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	jsonEncoder(w, resp)
}

func jsonEncoder(w http.ResponseWriter, resp interface{}) {
	e := json.NewEncoder(w).Encode(resp)
	if e != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
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

	if grpcMessage == "an internal error occurred" {
		return i18n.Message(errors.New("internal"), lang)
	}

	return grpcMessage
}
