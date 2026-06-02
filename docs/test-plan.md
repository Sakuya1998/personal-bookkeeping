# 测试计划

> 文档版本: v4.0 | 最后更新: 2026-06-02

---

## 1. 测试策略

### 1.1 测试金字塔

```
         ╱─────╲
        ╱  E2E  ╲          ═ 关键路径的端到端测试（Docker Compose 环境）
       ╱─────────╲
      ╱ 集成测试  ╲        ═ Handler + DB 集成测试
     ╱─────────────╲
    ╱   单元测试    ╲      ═ cache / queue / task / service / middleware / otel / config
   ╱─────────────────╲
  ╱  static analysis  ╲    ═ go vet + govulncheck + eslint + staticcheck
 ╱─────────────────────╲
```

### 1.2 现有测试覆盖

| 包 | 测试文件 | 行数 | 质量 |
|----|---------|------|------|
| `handler` | `handler_test.go` | 1,029 行 | ✅ 集成测试, 覆盖全 CRUD + Analytics + Auth + Ledger/Transaction 端点 |
| `handler` | `handler_sprint2_test.go` | 344 行 | ✅ Sprint 2 新增功能 (批量操作/导出/日历/标签/设置) |
| `handler` | `handler_sprint3_test.go` | 570 行 | ✅ 31 个测试全部填充断言 |
| `middleware` | `auth_test.go` | 217 行 | ✅ JWT 解析、Token 黑名单、Authorization header 检查 |
| `service` | `exchange_test.go` | 423 行 | ✅ 汇率正向/反向计算、缓存读写、DB miss 降级 |
| `service` | `ocr_test.go` | — | ✅ OCR 提取逻辑 + mock HTTP 测试 |
| `service` | `report_test.go` | — | ✅ parsePeriod/calcChange/truncate 纯函数测试 |
| `service` | `rate_updater_test.go` | — | ✅ httpGet/storeRates/fetchRates 测试 |
| `service` | `auth_test.go` | 139 行 | ✅ 令牌生成/验证/黑名单/缓存降级 (6 用例) |
| `service` | `ledger_test.go` | 201 行 | ✅ SplitTags/CSVRow/CSVHeader/FormatAmount/stringsJoin (13 用例) |
| `service` | `recurring_test.go` | 172 行 | ✅ ComputeNextRunDate 纯函数测试 (15 用例) |
| `task` | `tasks_test.go` | 603 行 | ✅ CSV/JSON 导入导出 + 调度器 mock 测试 |
| `cache/memory` | `memory_test.go` | 182 行 | ✅ 100% 纯内存，独立可并行 |
| `cache/tiered` | `tiered_test.go` | 211 行 | ✅ 纯内存，含故障降级 |
| `cache/edge` | `cache_edge_test.go` | 278 行 | ✅ 边缘场景：并发、过期、淘汰 |
| `queue` | `queue_test.go` | 135 行 | ✅ interface mock + 标准用例 |
| `queue` | `queue_edge_test.go` | 126 行 | ✅ 边缘场景：超时、空队列、故障恢复 |
| `otel` | `otel_test.go` | 295 行 | ✅ Init/Shutdown/Middleware/Metrics/Meter 接口 (14 用例) |
| `config` | `config_test.go` | 72 行 | ✅ DSN/L1 缓存时长/默认值 (6 用例) |
| `infra/middleware` | `ratelimit_test.go` | 99 行 | ✅ 滑动窗口/并发/突发/零速率 (7 用例) |
| `frontend/store` | `appStore.test.ts` | 244 行 | ✅ Zustand 状态管理测试 |
| `frontend/api` | `client.test.ts` | 153 行 | ✅ Axios 客户端、请求/响应拦截器 |
| `frontend/utils` | `currency.test.ts` | 54 行 | ✅ 货币格式化、金额计算 |

### 1.3 缺口分析

| 区域 | 现状 | 风险 |
|------|------|------|
| Handler 层 | v1/v2 CRUD + Analytics + Auth 完整覆盖 (1,029+344 行) | ✅ 已闭环 |
| Handler Sprint 3 | 31 个测试全部填充断言 (21 完整→31 完整) | ✅ 已闭环 |
| Service 层 | 汇率 + OCR + 报表 + 汇率更新 + auth + ledger + recurring (7 个测试文件) | ✅ 已闭环 |
| Task 层 | CSV/JSON/调度器全部覆盖 (603 行, 含 goroutine + 去重测试) | ✅ 已闭环 |
| Middleware | JWT/Token 黑名单/Header 检查 (217 行) + 速率限制 (99 行) | ✅ 已闭环 |
| OTEL 可观测性 | Init/Shutdown/Middleware/Metrics/Meter (295 行, 14 用例) | ✅ 已闭环 |
| 配置加载 | DSN/L1 缓存时长/默认值 (72 行, 6 用例) | ✅ 已闭环 |
| 前端 | Zustand + Axios + 货币工具 + BudgetPage + RecurringPage (5 文件, 47 测试) | ✅ S3 页面组件测试已补齐 |
| CI | 基本 lint + build + 覆盖率门禁 + govulncheck + staticcheck | ✅ 已增强 |
| E2E | 无 E2E 测试 | ⚠️ 低风险, 核心功能通过集成测试覆盖 |

---

## 2. 测试范围与优先级

### P0 — 核心业务逻辑（必须）

| 模块 | 测试类型 | 覆盖内容 |
|------|---------|---------|
| Transaction Create/Update | 单元 + 集成 | 金额计算、汇率折算、本位币计算、tags 格式化 |
| Auth Register/Login | 集成 | 用户名重复、密码哈希、JWT 签发、token 黑名单 |
| Ledger Summary | 集成 | 收入/支出汇总、分类支出排行 SQL |
| Category List | 集成 | 树形组装、缓存命中/失效 |
| Exchange Rate | 集成 | 反向汇率、缓存读写、CREATE OR UPDATE |

### P1 — 边界与异常路径

| 模块 | 测试类型 | 覆盖内容 |
|------|---------|---------|
| Transaction | 单元 | amount=0（拒绝）、金额边界、大浮点数精度、空 tags、过长 description |
| Auth | 单元 | 密码长度不足、用户名过短、email 格式错误 |
| Ledger Delete | 集成 | 级联删除 transaction + category |
| Category Delete | 集成 | 有关联交易时拒绝删除 |
| Pagination | 集成 | page < 1、pageSize > 100、0 数据、MAX INT |
| Cache Miss | 单元 | 不存在的 key 返回 ErrMiss |
| Tiered L2 故障 | 单元 | L2 不可用时 L1 降级 |

### P2 — Service 与 Task 层

| 模块 | 测试类型 | 覆盖内容 |
|------|---------|---------|
| GetExchangeRate | 单元 | 正向汇率、反向汇率(倒数)、缓存命中、DB miss 返回 0 |
| ExportReport | 单元 | CSV 格式、JSON 格式、空数据、SQL 注入(参数化) |
| ImportTransactions | 单元 | CSV 解析、JSON 解析、批量写入(分批)、格式不支持 |
| Cache invalidation | 单元 | Create/Update/Delete 后缓存清除 |

### P3 — 安全与可观测性

| 模块 | 测试类型 | 覆盖内容 |
|------|---------|---------|
| JWT Auth | 单元 | 过期 token、篡改 token、缺失 Authorization header |
| Logout | 集成 | Token 加入黑名单、黑名单 token 被拒绝 |
| SQL injection | 静态分析 | 所有用户输入使用参数化查询 |
| CORS | 集成 | OPTIONS preflight、Origin 校验 |
| OTEL Init/Shutdown | 单元 | nil config、disabled、enabled with prometheus/stdout、无效 exporter |
| OTEL Middleware | 单元 | nil meter、请求指标记录 |
| OTEL Metrics | 单元 | Prometheus handler、metric 命名存在性、Meter 接口一致性 |

---

## 3. 测试基础设施

### 3.1 后端

```
backend/
├── internal/
│   ├── app/
│   │   ├── handler/
│   │   │   ├── handler_test.go              # 集成测试: 全 CRUD + Analytics + Auth (1,029 行)
│   │   │   ├── handler_sprint2_test.go       # Sprint 2 功能测试 (344 行)
│   │   │   ├── handler_sprint3_test.go       # Sprint 3 测试 (570 行, 含 10 个 stub)
│   │   │   └── testutil_test.go              # 测试辅助工具
│   │   ├── middleware/
│   │   │   └── auth_test.go                  # JWT/Token 黑名单 (217 行)
│   │   ├── service/
│   │   │   ├── exchange_test.go              # 汇率服务 (423 行)
│   │   │   ├── ocr_test.go                   # OCR 提取 + mock HTTP
│   │   │   ├── report_test.go                # 报表纯函数
│   │   │   ├── rate_updater_test.go           # 汇率更新 mock
│   │   │   ├── auth_test.go                  # Auth 服务 (139 行, 6 用例) ✨ v4.0
│   │   │   ├── ledger_test.go                # 账本工具函数 (201 行, 13 用例) ✨ v4.0
│   │   │   └── recurring_test.go             # 周期性交易 (172 行, 15 用例) ✨ v4.0
│   │   └── task/
│   │       └── tasks_test.go                 # 导入导出 + 调度器 mock (603 行)
│   └── infra/
│       ├── cache/
│       │   ├── memory_test.go                # 内存缓存 (182 行)
│       │   ├── tiered_test.go                # 分层缓存 (211 行)
│       │   └── cache_edge_test.go            # 缓存边缘场景 (278 行)
│       ├── queue/
│       │   ├── queue_test.go                 # 队列标准用例 (135 行)
│       │   └── queue_edge_test.go            # 队列边缘场景 (126 行)
│       ├── otel/
│       │   └── otel_test.go                  # OTEL 可观测性 (295 行, 14 用例) ✨ v4.0
│       ├── config/
│       │   └── config_test.go                # 配置测试 (72 行, 6 用例) ✨ v4.0
│       └── middleware/
│           └── ratelimit_test.go             # 速率限制 (99 行, 7 用例) ✨ v4.0
```

### 3.2 前端

```
frontend/
├── src/
│   └── __tests__/
│       ├── utils/currency.test.ts            # 货币工具函数 (54 行)
│       ├── store/appStore.test.ts            # Zustand 状态管理 (244 行)
│       └── api/client.test.ts                # Axios 客户端 (153 行)
```

### 3.3 工具链

| 工具 | 用途 | 后端/前端 |
|------|------|----------|
| `go test -count=1 -race` | 并发安全检测 | 后端 |
| `go test -coverprofile=coverage.out` | 覆盖率收集 | 后端 |
| `go test -short` | 跳过依赖外部服务的测试 | 后端 |
| `BOOKKEEPING_TEST_DSN` | 测试数据库 DSN | 后端 |
| `vitest` | JavaScript 测试框架 | 前端 |
| `@testing-library/react` | React 组件测试 | 前端 |
| `msw` | API Mock | 前端 |
| `staticcheck` | Go 静态分析 | 后端 |
| `govulncheck` | Go 漏洞检测 | 后端 |

---

## 4. 测试规范

### 4.1 命名规范

```
Go:
  func TestModule_Feature_Scenario(t *testing.T)
  例: TestTransaction_Create_MultiCurrencyConversion

TypeScript:
  describe('Module', () => {
    it('should do X when Y', () => { ... })
  })
```

### 4.2 断言规范

- Go: 使用标准库 `testing` 包 + 表格驱动测试
- 错误信息必须说明预期和实际值: `t.Fatalf("got %v, want %v", got, want)`
- 条件: `if got != want { t.Errorf(...) }` ，Fatal 仅在无法继续时使用

### 4.3 隔离规则

```
handler_test.go 数据库测试:
  - 每个测试使用唯一的用户名前缀 (t.Name())
  - 不依赖其他测试的执行顺序
  - 不依赖测试间共享数据
  - 失败不留下脏数据

cache / queue / service 纯逻辑测试:
  - 使用 t.Parallel()
  - 不依赖外部服务（Redis / DB / Kafka）
  - 使用短超时避免 sleep 等待
```

### 4.4 Skip 策略

```go
if testing.Short() {
  t.Skip("skipping DB-dependent test in short mode")
}
```

---

## 5. 执行计划

### Phase 1 — 测试基础设施（已完成）
- [x] 本文档
- [x] testutil 辅助函数
- [x] vitest 配置
- [x] Makefile test 目标增强

### Phase 2 — P0 核心测试（已完成）
- [x] Handler 层所有 CRUD 测试补全
- [x] Auth 测试（含 token 黑名单）
- [x] Transaction 金额计算测试
- [x] Ledger Summary / Analytics 测试

### Phase 3 — P1 边界测试（已完成）
- [x] 输入校验边界测试
- [x] 并发写入测试
- [x] 缓存失效场景

### Phase 4 — P2 Service + Task（已完成）
- [x] GetExchangeRate 反向汇率
- [x] CSV/JSON 导入导出
- [x] 前台 store + utils 测试

### Phase 5 — CI 增强（已完成）
- [x] Go 测试 CI job (go test -race)
- [x] 前端测试 CI job (vitest)
- [x] 覆盖率报告门禁 (≥ 30%)
- [x] Docker layer caching (Buildx + actions/cache)
- [x] 前端 tsc 类型检查
- [x] Makefile 统一入口

### Phase 6 — v3.0 测试补全（已闭环）
- [x] 填充 7 个空 stub (见 §6)
- [x] 补全 3 个缺断言测试 (见 §6)
- [x] Service 层: OCR 单元测试 (ocr_test.go, 9 用例)
- [x] Service 层: 报表纯函数测试 (report_test.go, 26 用例)
- [x] Service 层: 汇率自动更新测试 (rate_updater_test.go, 17 用例)
- [x] Task 层: 调度器 goroutine + 去重逻辑测试
- [x] 前端: BudgetPage 组件测试 (6 用例)
- [x] 前端: RecurringPage 组件测试 (7 用例)
- [ ] E2E 测试 (Docker Compose) — 待后续

### Phase 7 — v4.0 测试补全（已闭环）
- [x] OTEL 可观测性测试 (otel_test.go, 14 用例: Init/Shutdown/Middleware/Metrics/Meter)
- [x] Service 层: Auth 服务测试 (auth_test.go, 6 用例: 令牌生成/验证/黑名单/缓存降级)
- [x] Service 层: 账本工具函数测试 (ledger_test.go, 13 用例: CSVRow/FormatAmount/SplitTags/stringsJoin)
- [x] Service 层: 周期性交易测试 (recurring_test.go, 15 用例: ComputeNextRunDate 多频率)
- [x] 配置测试 (config_test.go, 6 用例: DSN/L1 缓存时长/默认值)
- [x] 速率限制中间件测试 (ratelimit_test.go, 7 用例: 滑动窗口/并发/突发/零速率)
- [x] 代码审查回归测试: 31 个 Code Review 问题对应测试增强
- [x] 缓存空值哨兵保护测试
- [x] golang-migrate 迁移测试 (migrate up/down 正确性)
- [ ] E2E 测试 (Docker Compose) — 待后续

---

## 6. Sprint 3 测试 stub 清单（已修复）

`backend/internal/app/handler/handler_sprint3_test.go` 中的 10 个未完成测试已于 **2026-05-27** 全部填充断言。`go test ./internal/app/handler/` 不再有 stub 警告。

### 其他未覆盖模块（已闭环）

以下模块的测试已全部填补，`go test ./...` 全绿无 stub：

| 模块 | 文件 | 状态 |
|------|------|------|
| Sprint 3 Handler | `handler_sprint3_test.go` | ✅ 31/31 完整断言 |
| OCR 服务 | `service/ocr_test.go` | ✅ mock HTTP + 正则提取 |
| 报表服务 | `service/report_test.go` | ✅ 纯函数测试 |
| 汇率更新 | `service/rate_updater_test.go` | ✅ mock HTTP + DB 逻辑 |
| 调度器 | `task/scheduler.go` | ✅ mock queue + goroutine + 去重 |
| 前端 BudgetPage | `BudgetPage.test.tsx` | ✅ 6 用例: 骨架屏/数据展示/弹窗/提交/删除 |
| 前端 RecurringPage | `RecurringPage.test.tsx` | ✅ 7 用例: 加载/空态/渲染/弹窗/编辑/删除 |
| Auth 服务 | `service/auth_test.go` | ✅ 6 用例: 令牌生成/验证/黑名单 ✨ v4.0 |
| 账本工具 | `service/ledger_test.go` | ✅ 13 用例: CSVRow/FormatAmount ✨ v4.0 |
| 周期性交易 | `service/recurring_test.go` | ✅ 15 用例: ComputeNextRunDate ✨ v4.0 |
| 配置模块 | `config/config_test.go` | ✅ 6 用例: DSN/缓存时长 ✨ v4.0 |
| OTEL 可观测性 | `otel/otel_test.go` | ✅ 14 用例: Init/Shutdown/Middleware ✨ v4.0 |
| 速率限制 | `middleware/ratelimit_test.go` | ✅ 7 用例: 滑动窗口/并发 ✨ v4.0 |

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| Handler 测试依赖真实 PG | CI 不稳定 | 添加 `-short` flag 跳过；用 Go test DB container |
| 浮点数精度 | 断言失败 | 使用 `math.Abs(got-want) < 1e-6` 近似比较 |
| 缓存测试时序 | flaky test | 使用 `time.NewTimer` 而非 `time.Sleep`；mock 时间 |
| 前端 Ant Design 版本 | 快照测试易碎 | 不用快照测试，用逻辑断言 |
| OTEL 全局状态污染 | 测试间干扰 | 使用 resetOtel() 重置全局状态 (见 otel_test.go) |
