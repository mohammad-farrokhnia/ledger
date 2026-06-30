package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mohammad-farrokhnia/ledger/internal/i18n"
)

var (
	appName string
	version string
)

func Init(name, ver string) {
	appName = name
	version = ver
}

type Meta struct {
	AppName     string `json:"appName"`
	Version     string `json:"version"`
	RequestID   string `json:"requestId"`
	Timestamp   string `json:"timestamp"`
	MessageCode string `json:"messageCode"`
	Message     string `json:"message"`
}

func newMeta(requestID, acceptLang string, code i18n.MessageCode) Meta {
	return Meta{
		AppName:     appName,
		Version:     version,
		RequestID:   requestID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		MessageCode: string(code),
		Message:     i18n.Translate(acceptLang, code),
	}
}

func OK(w http.ResponseWriter, r *http.Request, data any, code i18n.MessageCode) {
	write(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": newMeta(requestID(r), r.Header.Get("Accept-Language"), code),
	})
}

func Created(w http.ResponseWriter, r *http.Request, data any, code i18n.MessageCode) {
	write(w, http.StatusCreated, map[string]any{
		"data": data,
		"meta": newMeta(requestID(r), r.Header.Get("Accept-Language"), code),
	})
}

func Error(w http.ResponseWriter, r *http.Request, status int, code i18n.MessageCode) {
	write(w, status, map[string]any{
		"error": map[string]any{},
		"meta":  newMeta(requestID(r), r.Header.Get("Accept-Language"), code),
	})
}

func write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "response encode error", http.StatusInternalServerError)
	}
}

func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	return "—"
}
