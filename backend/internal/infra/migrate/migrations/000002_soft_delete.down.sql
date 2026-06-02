-- 000002_soft_delete.down.sql
-- Remove soft delete columns and restore simple unique indexes.

-- users: drop partial indexes, recreate simple unique indexes
DROP INDEX IF EXISTS idx_users_username;
CREATE UNIQUE INDEX idx_users_username ON users (username);
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON users (email);

-- exchange_rates
DROP INDEX IF EXISTS idx_exchange_rate_pair;
CREATE UNIQUE INDEX idx_exchange_rate_pair ON exchange_rates (from_currency, to_currency);

-- ledger_members
DROP INDEX IF EXISTS idx_ledger_user;
CREATE UNIQUE INDEX idx_ledger_user ON ledger_members (ledger_id, user_id);

-- drop deleted_at indexes and columns
DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_ledgers_deleted_at;
ALTER TABLE ledgers DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_categories_deleted_at;
ALTER TABLE categories DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_transactions_deleted_at;
ALTER TABLE transactions DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_exchange_rates_deleted_at;
ALTER TABLE exchange_rates DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_recurring_rules_deleted_at;
ALTER TABLE recurring_rules DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_budgets_deleted_at;
ALTER TABLE budgets DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_ledger_members_deleted_at;
ALTER TABLE ledger_members DROP COLUMN IF EXISTS deleted_at;
