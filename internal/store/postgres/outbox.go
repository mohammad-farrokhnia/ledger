package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mohammad-farrokhnia/go-ledger/internal/ledger"
)

func (s *Store) PollAuditOutbox(ctx context.Context, limit int) ([]ledger.AuditOutboxEntry, error) {
	const q = `
		SELECT id, action_type, payload, retry_count, max_retries, created_at
		FROM audit_outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres: poll audit outbox: %w", err)
	}
	defer rows.Close()

	var entries []ledger.AuditOutboxEntry
	for rows.Next() {
		var e ledger.AuditOutboxEntry
		var rawPayload []byte

		if err = rows.Scan(&e.ID, &e.ActionType, &rawPayload, &e.RetryCount, &e.MaxRetries, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan audit outbox row: %w", err)
		}

		if err = json.Unmarshal(rawPayload, &e.Payload); err != nil {
			return nil, fmt.Errorf("postgres: unmarshal audit payload: %w", err)
		}

		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (s *Store) MarkAuditEntrySent(ctx context.Context, id string) error {
	const q = `
		UPDATE audit_outbox
		SET status = 'SENT', processed_at = NOW()
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, q, id)
	return err
}

func (s *Store) MarkAuditEntryFailed(ctx context.Context, id, errMsg string) error {
	const q = `
		UPDATE audit_outbox
		SET
			retry_count = retry_count + 1,
			last_error  = $2,
			status = CASE
				WHEN retry_count + 1 >= max_retries THEN 'DEAD_LETTERED'::audit_status
				ELSE 'PENDING'::audit_status
			END,
			processed_at = CASE
				WHEN retry_count + 1 >= max_retries THEN NOW()
				ELSE NULL
			END
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, q, id, errMsg)
	return err
}

var _ ledger.OutboxStore = (*Store)(nil)
