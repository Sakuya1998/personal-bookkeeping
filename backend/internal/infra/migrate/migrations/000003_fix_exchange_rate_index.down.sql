-- 000003_fix_exchange_rate_index.down.sql
-- 回退到部分索引

DROP INDEX IF EXISTS idx_exchange_rate_pair;
CREATE UNIQUE INDEX idx_exchange_rate_pair ON exchange_rates (from_currency, to_currency) WHERE deleted_at IS NULL;
