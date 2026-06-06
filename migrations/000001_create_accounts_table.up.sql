CREATE TYPE account_type AS ENUM ('USER', 'SYSTEM', 'PROVIDER');

CREATE TABLE accounts (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    type          account_type NOT NULL DEFAULT 'USER',

    currency_code CHAR(3)     NOT NULL,

    balance       BIGINT      NOT NULL DEFAULT 0,


    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_currency_code ON accounts (currency_code);

CREATE INDEX idx_accounts_type ON accounts (type);

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();
