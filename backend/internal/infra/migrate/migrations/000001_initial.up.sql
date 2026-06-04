-- 000001_initial.up.sql
-- Initial schema for Personal Bookkeeping app

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id           UUID        PRIMARY KEY,
    username     VARCHAR(50) NOT NULL,
    email        VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- ============================================================
-- ledgers
-- ============================================================
CREATE TABLE IF NOT EXISTS ledgers (
    id            UUID         PRIMARY KEY,
    user_id       UUID         NOT NULL,
    name          VARCHAR(100) NOT NULL,
    description   VARCHAR(500),
    base_currency VARCHAR(10)  NOT NULL DEFAULT 'CNY',
    icon          VARCHAR(50),
    color         VARCHAR(20),
    is_archived   BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order    INTEGER      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ledgers_user_id ON ledgers (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ledgers_user_name ON ledgers (user_id, name);

-- ============================================================
-- categories
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id         UUID         PRIMARY KEY,
    user_id    UUID         NOT NULL,
    ledger_id  UUID,
    name       VARCHAR(50)  NOT NULL,
    type       VARCHAR(10)  NOT NULL,
    icon       VARCHAR(50),
    color      VARCHAR(20),
    parent_id  UUID,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_category_type CHECK (type IN ('income', 'expense'))
);

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories (user_id);
CREATE INDEX IF NOT EXISTS idx_categories_ledger_id ON categories (ledger_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_ledger_name_type ON categories (ledger_id, name, type);

-- ============================================================
-- transactions
-- ============================================================
CREATE TABLE IF NOT EXISTS transactions (
    id               UUID          PRIMARY KEY,
    ledger_id        UUID          NOT NULL,
    user_id          UUID          NOT NULL,
    category_id      UUID          NOT NULL,
    type             VARCHAR(10)   NOT NULL,
    amount           DECIMAL(18,2) NOT NULL,
    currency         VARCHAR(10)   NOT NULL DEFAULT 'CNY',
    exchange_rate    DECIMAL(18,8) NOT NULL DEFAULT 1.0,
    base_amount      DECIMAL(18,2) NOT NULL,
    description      TEXT,
    transaction_date DATE          NOT NULL,
    tags             TEXT,
    is_reconciled    BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_transaction_type CHECK (type IN ('income', 'expense'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_ledger_id ON transactions (ledger_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions (category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_date ON transactions (transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_ledger_user_date ON transactions (ledger_id, user_id, transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_user_type ON transactions (user_id, type);

-- ============================================================
-- exchange_rates
-- ============================================================
CREATE TABLE IF NOT EXISTS exchange_rates (
    id            UUID          PRIMARY KEY,
    from_currency VARCHAR(10)   NOT NULL,
    to_currency   VARCHAR(10)   NOT NULL,
    rate          DECIMAL(18,8) NOT NULL,
    date          DATE          NOT NULL,
    source        VARCHAR(50),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exchange_rate_pair ON exchange_rates (from_currency, to_currency);

-- ============================================================
-- recurring_rules
-- ============================================================
CREATE TABLE IF NOT EXISTS recurring_rules (
    id            UUID          PRIMARY KEY,
    user_id       UUID          NOT NULL,
    ledger_id     UUID          NOT NULL,
    category_id   UUID          NOT NULL,
    type          VARCHAR(10)   NOT NULL,
    amount        DECIMAL(18,2) NOT NULL,
    currency      VARCHAR(10)   NOT NULL DEFAULT 'CNY',
    description   TEXT,
    tags          TEXT,
    frequency     VARCHAR(20)   NOT NULL,
    interval      INTEGER       NOT NULL DEFAULT 1,
    day_of_month  INTEGER,
    weekday       INTEGER,
    start_date    DATE          NOT NULL,
    end_date      DATE,
    next_run_date DATE          NOT NULL,
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_recurring_type CHECK (type IN ('income', 'expense'))
);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_user_id ON recurring_rules (user_id);
CREATE INDEX IF NOT EXISTS idx_recurring_rules_ledger_id ON recurring_rules (ledger_id);
CREATE INDEX IF NOT EXISTS idx_recurring_rules_next_run_date ON recurring_rules (next_run_date);

-- ============================================================
-- budgets
-- ============================================================
CREATE TABLE IF NOT EXISTS budgets (
    id          UUID          PRIMARY KEY,
    user_id     UUID          NOT NULL,
    ledger_id   UUID          NOT NULL,
    category_id UUID,
    month       VARCHAR(7)    NOT NULL,
    amount      DECIMAL(18,2) NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_budget_month_format CHECK (month ~ '^[0-9]{4}-[0-9]{2}$')
);

CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets (user_id);
CREATE INDEX IF NOT EXISTS idx_budgets_ledger_id ON budgets (ledger_id);
CREATE INDEX IF NOT EXISTS idx_budgets_category_id ON budgets (category_id);

-- ============================================================
-- ledger_members
-- ============================================================
CREATE TABLE IF NOT EXISTS ledger_members (
    id         UUID        PRIMARY KEY,
    ledger_id  UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    role       VARCHAR(20) NOT NULL DEFAULT 'member',
    invited_by UUID,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_ledger_member_role CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ledger_user ON ledger_members (ledger_id, user_id);
