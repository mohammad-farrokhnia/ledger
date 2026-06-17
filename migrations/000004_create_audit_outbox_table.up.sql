-- audit_outbox implements the transactional outbox pattern.
-- Every CreateTransaction inserts a row here WITHIN THE SAME DB TRANSACTION.
-- This guarantees: if the ledger transaction commits, the audit record exists.
-- A background worker reads PENDING rows and ships them to the webhook.
-- If the webhook is down, rows stay PENDING and are retried — no audit is lost.

CREATE TYPE audit_status AS ENUM ('PENDING', 'SENT', 'FAILED', 'DEAD_LETTERED');

CREATE TABLE audit_outbox (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    action_type   VARCHAR(100) NOT NULL,
    payload       JSONB        NOT NULL,
    status        audit_status NOT NULL DEFAULT 'PENDING',
    retry_count   INT          NOT NULL DEFAULT 0,
    max_retries   INT          NOT NULL DEFAULT 3,
    last_error    TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ
);

-- Partial index — only scans PENDING rows, which is the hot path.
CREATE INDEX idx_audit_outbox_pending ON audit_outbox (created_at)
    WHERE status = 'PENDING';
