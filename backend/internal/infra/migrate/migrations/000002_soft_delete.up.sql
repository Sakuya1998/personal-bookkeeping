-- 000002_soft_delete.up.sql
-- Add soft delete (deleted_at) to all tables.
-- Recreate unique indexes as partial indexes (WHERE deleted_at IS NULL)
-- so soft-deleted records don't block re-creation of the same name/pair.

-- ============================================================
-- users
-- ============================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

DROP INDEX IF EXISTS idx_users_username;
CREATE UNIQUE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;

-- ============================================================
-- ledgers
-- ============================================================
ALTER TABLE ledgers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_ledgers_deleted_at ON ledgers (deleted_at);

-- ============================================================
-- categories
-- ============================================================
ALTER TABLE categories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories (deleted_at);

-- ============================================================
-- transactions
-- ============================================================
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions (deleted_at);

-- ============================================================
-- exchange_rates
-- ============================================================
ALTER TABLE exchange_rates ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_exchange_rates_deleted_at ON exchange_rates (deleted_at);

DROP INDEX IF EXISTS idx_exchange_rate_pair;
CREATE UNIQUE INDEX idx_exchange_rate_pair ON exchange_rates (from_currency, to_currency) WHERE deleted_at IS NULL;

-- ============================================================
-- recurring_rules
-- ============================================================
ALTER TABLE recurring_rules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_recurring_rules_deleted_at ON recurring_rules (deleted_at);

-- ============================================================
-- budgets
-- ============================================================
ALTER TABLE budgets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_budgets_deleted_at ON budgets (deleted_at);

-- ============================================================
-- ledger_members
-- ============================================================
ALTER TABLE ledger_members ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_ledger_members_deleted_at ON ledger_members (deleted_at);

DROP INDEX IF EXISTS idx_ledger_user;
CREATE UNIQUE INDEX idx_ledger_user ON ledger_members (ledger_id, user_id) WHERE deleted_at IS NULL;
