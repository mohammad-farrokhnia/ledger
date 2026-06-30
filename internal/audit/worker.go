package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mohammad-farrokhnia/ledger/internal/ledger"
)

type Worker struct {
	store    ledger.OutboxStore
	hookURL  string
	client   *http.Client
	interval time.Duration
}

func NewWorker(store ledger.OutboxStore, hookURL string, interval time.Duration) *Worker {
	return &Worker{
		store:    store,
		hookURL:  hookURL,
		client:   &http.Client{Timeout: 5 * time.Second},
		interval: interval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("audit outbox worker started", "interval", w.interval.String(), "hook_url", w.hookURL)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("audit outbox worker stopping")
			return
		case <-ticker.C:
			w.flush(ctx)
		}
	}
}

func (w *Worker) flush(ctx context.Context) {
	entries, err := w.store.PollAuditOutbox(ctx, 100)
	if err != nil {
		slog.Error("audit outbox: poll failed", "error", err)
		return
	}

	for _, entry := range entries {
		if err = w.deliver(ctx, entry); err != nil {
			slog.Warn("audit outbox: delivery failed",
				"id", entry.ID,
				"action", entry.ActionType,
				"retry_count", entry.RetryCount,
				"error", err,
			)
			if markErr := w.store.MarkAuditEntryFailed(ctx, entry.ID, err.Error()); markErr != nil {
				slog.Error("audit outbox: mark failed error", "error", markErr)
			}
			continue
		}

		slog.Debug("audit outbox: delivered", "id", entry.ID, "action", entry.ActionType)
		if markErr := w.store.MarkAuditEntrySent(ctx, entry.ID); markErr != nil {
			slog.Error("audit outbox: mark sent error", "error", markErr)
		}
	}
}

func (w *Worker) deliver(ctx context.Context, entry ledger.AuditOutboxEntry) error {
	body, err := json.Marshal(map[string]any{
		"action_type": entry.ActionType,
		"payload":     entry.Payload,
		"created_at":  entry.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.hookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	closeResponseBody(resp)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("hook returned %d", resp.StatusCode)
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
