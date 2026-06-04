-- 000004_add_missing_constraints.up.sql
-- 补充初始 SQL 中遗漏的约束

-- 1. ledgers: 同一用户下账本名唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_ledgers_user_name ON ledgers (user_id, name);

-- 2. categories: 同一账本下分类名+类型唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_ledger_name_type ON categories (ledger_id, name, type);

-- 3. ledger_members: role 值校验
ALTER TABLE ledger_members ADD CONSTRAINT ck_ledger_member_role CHECK (role IN ('owner', 'admin', 'member'));

-- 4. budgets: month 格式校验 YYYY-MM
-- PostgreSQL 9.2+ 支持 SIMILAR TO，但直接用 LIKE 更简洁
ALTER TABLE budgets ADD CONSTRAINT ck_budget_month_format CHECK (month ~ '^[0-9]{4}-[0-9]{2}$');
