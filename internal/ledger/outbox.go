package ledger

import (
	"context"
	"time"
)

type AuditOutboxEntry struct {
	ID         string
	ActionType string
	Payload    map[string]any
	RetryCount int
	MaxRetries int
	CreatedAt  time.Time
}

type OutboxStore interface {
	PollAuditOutbox(ctx context.Context, limit int) ([]AuditOutboxEntry, error)
	MarkAuditEntrySent(ctx context.Context, id string) error
	MarkAuditEntryFailed(ctx context.Context, id, errMsg string) error
}
