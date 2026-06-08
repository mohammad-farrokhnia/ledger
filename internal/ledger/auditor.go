package ledger

import "context"

type AuditEntry struct {
	ActionType string
	Payload    map[string]any
}

type Auditor interface {
	Log(ctx context.Context, entry AuditEntry) error
}
