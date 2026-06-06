CREATE TYPE transaction_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED');

CREATE TABLE transactions (
    id               UUID               PRIMARY KEY DEFAULT gen_random_uuid(),

    idempotency_key  UUID               NOT NULL UNIQUE,

    from_account_id  UUID               NOT NULL REFERENCES accounts(id),
    to_account_id    UUID               NOT NULL REFERENCES accounts(id),

    amount           BIGINT             NOT NULL,

    currency_code    CHAR(3)            NOT NULL,

    status           transaction_status NOT NULL DEFAULT 'PENDING',

    metadata         JSONB,

    created_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_different_accounts CHECK (from_account_id != to_account_id),

    CONSTRAINT chk_positive_amount CHECK (amount > 0)
);

CREATE INDEX idx_transactions_from_account  ON transactions (from_account_id);
CREATE INDEX idx_transactions_to_account    ON transactions (to_account_id);
CREATE INDEX idx_transactions_status        ON transactions (status);
CREATE INDEX idx_transactions_created_at    ON transactions (created_at DESC);

CREATE TRIGGER set_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();
