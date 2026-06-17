CREATE TABLE entries (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id     UUID        NOT NULL REFERENCES accounts(id),
    transaction_id UUID        NOT NULL REFERENCES transactions(id),

    amount         BIGINT      NOT NULL,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()

);

CREATE INDEX idx_entries_account_id     ON entries (account_id);

CREATE INDEX idx_entries_transaction_id ON entries (transaction_id);

CREATE INDEX idx_entries_created_at     ON entries (created_at DESC);
