CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    username varchar(50) NOT NULL UNIQUE,
    email varchar(100) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ledgers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    description varchar(500),
    base_currency varchar(10) NOT NULL DEFAULT 'CNY',
    icon varchar(50),
    color varchar(20),
    is_archived boolean NOT NULL DEFAULT false,
    sort_order int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_ledgers_user_id ON ledgers(user_id);

CREATE TABLE categories (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ledger_id uuid REFERENCES ledgers(id) ON DELETE SET NULL,
    name varchar(50) NOT NULL,
    type varchar(10) NOT NULL CHECK (type IN ('income', 'expense')),
    icon varchar(50),
    color varchar(20),
    parent_id uuid REFERENCES categories(id) ON DELETE SET NULL,
    sort_order int NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE INDEX idx_categories_ledger_id ON categories(ledger_id);

CREATE TABLE transactions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    ledger_id uuid NOT NULL REFERENCES ledgers(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    type varchar(10) NOT NULL CHECK (type IN ('income', 'expense')),
    amount decimal(18,2) NOT NULL,
    currency varchar(10) NOT NULL DEFAULT 'CNY',
    exchange_rate decimal(18,8) NOT NULL DEFAULT 1.0,
    base_amount decimal(18,2) NOT NULL,
    description text,
    transaction_date date NOT NULL,
    tags text,
    is_reconciled boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_transactions_ledger_id ON transactions(ledger_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);

CREATE TABLE exchange_rates (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency varchar(10) NOT NULL,
    to_currency varchar(10) NOT NULL,
    rate decimal(18,8) NOT NULL,
    date date NOT NULL,
    source varchar(50),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_exchange_rates_curr ON exchange_rates(from_currency, to_currency, date);
