-- 000001_initial.down.sql
-- Rollback initial schema

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS recurring_rules;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS ledger_members;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS ledgers;
DROP TABLE IF EXISTS users;
