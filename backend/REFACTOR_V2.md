# Backend Refactoring Plan v2 — Status ✅

## 最终状态

```
Handler (HTTP only)
  └─ svc.CreateTransaction(...)     ✅ 业务方法
  └─ svc.Cache / svc.Queue          ✅ 通过 DI
  └─ h.svc.DB.xxx / database.Get*   ❌ 零处（除 test 文件）
```

---

## Phase 1 — Handler → Service ✅

| Handler | 直接 DB 调用数 | 变更 |
|---------|---------------|------|
| `transaction.go` | ~15 → 0 | 6 方法全部迁移，-136 行 |
| `ledger.go` | ~20 → 0 | 12 方法全部迁移，-64 行 |
| `analytics.go` | ~5 → 0 | 3 方法全部迁移，-96 行 |
| `budget.go` | ~15 → 0 | 4 方法全部迁移，-142 行；移除包级 `CheckBudgetOverrun` |
| `category.go` | ~10 → 0 | 4 方法全部迁移，-101 行；缓存逻辑移到 service |
| `recurring.go` | ~10 → 0 | 5 方法全部迁移；`computeNextRunDate`/`daysInMonth` 移到 service |
| `exchange_rate.go` | ~8 → 0 | 4 方法全部迁移，-58 行 |
| `report.go` | ~4 → 0 | 2 方法全部迁移；`BuildReportData` 方法化 |
| `auth.go` | ~8 → 0 | 新建 `service/auth.go`，Register/Login/ChangePassword/ChangeEmail 全部迁移；移除 `bcrypt`/`jwt`/`uuid` 导入和包级辅助函数 |

## Phase 2 — Cache/Queue 归一 ✅

| 文件 | 旧调用 | 新调用 |
|------|--------|--------|
| `ledger.go:337` | `database.GetQueue()` | `h.svc.SubmitExportTask()` |
| `exchange_rate.go:173` | `database.GetCache()` | `s.Cache.Delete()` (service 内部) |
| `category.go:59, 294` | `database.GetCache()` | `s.Cache` (service 内部) |
| `auth.go:190` | `database.GetCache()` | `h.svc.Cache` |

## Phase 3 — Service 补充 ✅

| 文件 | 方法 |
|------|------|
| `service/auth.go` (新建) | Register / Login / ChangePassword / ChangeEmail / BlacklistToken |
| `service/recurring.go` (新建) | Create / Update / Delete / List / Upcoming |
| `service/exchange_rate.go` (新建) | List / Create / Latest / Delete + invalidateCache |

## Phase 4 — 清理 ✅

| 项 | 状态 |
|----|------|
| `model/` → `models/` 目录改名 | 完成 |
| 21 个 import 路径同步更新 | 完成 |
| `report.go` 死代码 `var _ = slog.Default` | 已移除 |
| Router 注入 Service | 已存在 |

## 总工作量

| Phase | 文件数 | 耗时 | 优先级 |
|-------|--------|------|--------|
| Phase 1 — Handler 迁移 | 8 handler 文件 | 完成 | P0 ✅ |
| Phase 2 — Cache/Queue 归一 | 4 文件 | 完成 | P1 ✅ |
| Phase 3 — Service 补充 | 2 新文件 | 完成 | P1 ✅ |
| Phase 4 — 清理 | 3 项 | 完成 | P2 ✅ |
| Phase 5 — Infra 层清理 | 6 项 | 完成 | P1 ✅ |

## Phase 5 — Infra 层清理 ✅

| 项 | 变更 |
|----|------|
| `ratelimit.go` 移入 `infra/middleware/` | 纯基础设施，零业务依赖 |
| `database.go` 移除 `app/models` 导入 | AutoMigrate 移至 `main.go` |
| `database.go` 移除 `cch`/`q` 全局变量 | `GetCache()`/`GetQueue()` 移至各自包 |
| `cache` 包新增 `SetDefault()`/`GetDefault()` | 替代 `database.GetCache()` |
| `queue` 包新增 `SetDefault()`/`GetDefault()` | 替代 `database.GetQueue()` |
| `exchange.go`/`rate_updater.go` 使用 DI | `database.GetDB()` → 构造函数注入 |

### 当前依赖方向

```
main.go
  ├─ infrà/database   (GetDB 初始化)
  ├─ infra/cache      (SetDefault + cache 实现)
  ├─ infra/queue      (SetDefault + queue 实现)
  └─ app/service      (业务逻辑，通过 DI 注入 DB/Cache/Queue)

app/service
  └─ infra/database.GetDB()
  └─ infra/cache.GetDefault()
  └─ infra/queue.GetDefault()

infra/ 各包之间：零依赖（database/cache/queue/logger/otel 互不引用）
infra/ → app/：零引用 ✅
```

## 最终状态

所有已知待办已完成，项目处于可部署状态。当前唯一的操作建议是将 `config.yaml.example` 中的占位符替换为实际值后复制为 `config.yaml`，并确保 `exchange_rate.api_key` 填写真实 API key。

## 现有测试覆盖

| 包 | 测试内容 | 状态 |
|----|----------|------|
| `infra/cache` | memory（Get/Set/Miss/Delete/Exists/Expiration/Flush/Update/TTL/Eviction/Concurrency/DefaultTTL）、tiered（L1/L2/Miss/WriteThrough/Delete/Exists/Flush/FailureDegradation）、edge cases（LargeValue/DeleteNonExistent/FlushEmpty/ExistsAfterExpiration/Concurrent） | ✅ |
| `infra/queue` | interface contract、lifecycle（SubmitBeforeStart/SubmitAfterStart/ShutdownRejects/DoubleStart/DoubleShutdown）、factory（disabled/unknown）、inmemory edge cases | ✅ |
| `infra/config` | DSN formatting、L1Duration logic（explicit/from TTL/small TTL/zero） | ✅ |
| `infra/middleware` | RateLimiter（AllowWithinLimit/RejectWhenExceeded/DifferentIPsIndependent/WindowSliding/ZeroRate/ConcurrentAccess/SingleBurst） | ✅ |
| `app/middleware` | JWT Claims generate/parse/expired/wrong signing method/blacklist、AuthRequired（missing header/invalid format/invalid token/expired token/blacklisted token/valid token） | ✅ |
| `app/handler` | Integration tests（E2E HTTP via real DB, 3 sprint suites） | ✅ |
| `app/service` | exchange（cache hit/parse error/reverse/not found/zero/table-driven）、ocr（success/HTTP error/empty/invalid JSON/known patterns）、report（parsePeriod/calcChange/truncate）、auth（generateToken/different users/wrong secret/BlacklistToken with cache/nil cache/unique keys）、recurring（ComputeNextRunDate 11 cases/DaysInMonth）、ledger（splitTags/trim/CSVRow/CSVHeader/FormatAmount/stringsJoin） | ✅ |
