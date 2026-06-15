// Package response provides the unified HTTP response envelope for go-ledger.
// Every response — success or error — uses the same {data, meta} shape.
// This mirrors the pattern from LibreCore's pkg/response package.
package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mohammad-farrokhnia/go-ledger/internal/i18n"
)

var (
	appName string
	version string
)

// Init sets the application name and version embedded in every Meta.
// Call once from main() before starting the server.
func Init(name, ver string) {
	appName = name
	version = ver
}

// Meta is present on every response.
// MessageCode lets clients do their own i18n if needed.
// Message is already translated to the caller's Accept-Language.
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

// OK writes a 200 success response.
func OK(w http.ResponseWriter, r *http.Request, data any, code i18n.MessageCode) {
	write(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": newMeta(requestID(r), r.Header.Get("Accept-Language"), code),
	})
}

// Created writes a 201 created response.
func Created(w http.ResponseWriter, r *http.Request, data any, code i18n.MessageCode) {
	write(w, http.StatusCreated, map[string]any{
		"data": data,
		"meta": newMeta(requestID(r), r.Header.Get("Accept-Language"), code),
	})
}

// Error writes an error response with the given HTTP status.
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
