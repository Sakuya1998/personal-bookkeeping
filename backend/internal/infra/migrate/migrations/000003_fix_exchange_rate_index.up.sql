-- 000003_fix_exchange_rate_index.up.sql
-- exchange_rates 表只做 upsert（INSERT ON CONFLICT），没有软删除必要。
-- ON CONFLICT 需要无条件唯一索引，把部分索引改回完整索引。

DROP INDEX IF EXISTS idx_exchange_rate_pair;
CREATE UNIQUE INDEX idx_exchange_rate_pair ON exchange_rates (from_currency, to_currency);
