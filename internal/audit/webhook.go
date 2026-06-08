package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

var (
	WebhookAuditorSyncMode  = "sync"
	WebhookAuditorAsyncMode = "async"
)

type WebhookAuditor struct {
	hookURL string
	mode    string
	client  *http.Client
}

func NewWebhookAuditor(hookURL, mode string) *WebhookAuditor {
	return &WebhookAuditor{
		hookURL: hookURL,
		mode:    mode,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (w *WebhookAuditor) Log(ctx context.Context, entry ledger.AuditEntry) error {
	if w.mode == "async" {
		go func() {
			if err := w.post(context.Background(), entry); err != nil {
				slog.Error("audit webhook failed",
					"action", entry.ActionType,
					"error", err,
				)
			}
		}()
		return nil
	}

	return w.post(ctx, entry)
}

func (w *WebhookAuditor) post(ctx context.Context, entry ledger.AuditEntry) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.hookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("audit: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("audit: post to hook: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("audit: hook returned status %d", resp.StatusCode)
	}

	return nil
}

func closeResponseBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	err := resp.Body.Close()
	if err != nil {
		slog.Error("audit: close response body", "error", err)
	}
}

var _ ledger.Auditor = (*WebhookAuditor)(nil)
