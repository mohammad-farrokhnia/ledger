DROP TRIGGER IF EXISTS set_accounts_updated_at ON accounts;
DROP FUNCTION IF EXISTS trigger_set_updated_at();
DROP TABLE IF EXISTS accounts;
DROP TYPE IF EXISTS account_type;
