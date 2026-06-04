-- 000004_add_missing_constraints.down.sql
-- 回退 000004 添加的约束

ALTER TABLE budgets DROP CONSTRAINT IF EXISTS ck_budget_month_format;
ALTER TABLE ledger_members DROP CONSTRAINT IF EXISTS ck_ledger_member_role;
DROP INDEX IF EXISTS idx_categories_ledger_name_type;
DROP INDEX IF EXISTS idx_ledgers_user_name;
