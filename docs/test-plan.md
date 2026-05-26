# 测试计划

> 文档版本: v2.0 | 最后更新: 2026-05-26

---

## 1. 测试策略

### 1.1 测试金字塔

```
         ╱─────╲
        ╱  E2E  ╲          ═ 关键路径的端到端测试（Docker Compose 环境）
       ╱─────────╲
      ╱ 集成测试  ╲        ═ Handler + DB 集成测试
     ╱─────────────╲
    ╱   单元测试    ╲      ═ cache / queue / task / service / middleware
   ╱─────────────────╲
  ╱  static analysis  ╲    ═ go vet + govulncheck + eslint
 ╱─────────────────────╲
```

### 1.2 现有测试覆盖

| 包 | 测试文件 | 行数 | 质量 |
|----|---------|------|------|
| `handler` | `handler_test.go` | 1,029 行 | ✅ 集成测试, 覆盖全 CRUD + Analytics + Auth + Ledger/Transaction 端点 |
| `handler` | `handler_sprint2_test.go` | 344 行 | ✅ Sprint 2 新增功能 (批量操作/导出/日历/标签/设置) |
| `middleware` | `auth_test.go` | 217 行 | ✅ JWT 解析、Token 黑名单、Authorization header 检查 |
| `service` | `exchange_test.go` | 423 行 | ✅ 汇率正向/反向计算、缓存读写、DB miss 降级 |
| `task` | `tasks_test.go` | 545 行 | ✅ CSV/JSON 导入导出、分批写入、格式校验 |
| `cache/memory` | `memory_test.go` | 182 行 | ✅ 100% 纯内存，独立可并行 |
| `cache/tiered` | `tiered_test.go` | 211 行 | ✅ 纯内存，含故障降级 |
| `cache/edge` | `cache_edge_test.go` | 278 行 | ✅ 边缘场景：并发、过期、淘汰 |
| `queue` | `queue_test.go` | 135 行 | ✅ interface mock + 标准用例 |
| `queue` | `queue_edge_test.go` | 126 行 | ✅ 边缘场景：超时、空队列、故障恢复 |
| `frontend/store` | `appStore.test.ts` | 244 行 | ✅ Zustand 状态管理测试 |
| `frontend/api` | `client.test.ts` | 153 行 | ✅ Axios 客户端、请求/响应拦截器 |
| `frontend/utils` | `currency.test.ts` | 54 行 | ✅ 货币格式化、金额计算 |

### 1.3 缺口分析

| 区域 | 现状 | 风险 |
|------|------|------|
| Handler 层 | 覆盖全 CRUD + Analytics + Auth + Sprint 2 功能 (1,029+344 行) | ✅ 已闭环, 持续关注新增功能 |
| Service 层 | 汇率计算单元测试覆盖 (exchange_test.go 423 行) | ✅ 已闭环 |
| Middleware 层 | JWT/Token 黑名单/Header 检查覆盖 (auth_test.go 217 行) | ✅ 已闭环 |
| Task 层 | CSV/JSON 导入导出批量处理覆盖 (tasks_test.go 545 行) | ✅ 已闭环 |
| 前端 | Zustand store / Axios client / 货币工具函数测试 (3 文件, 451 行) | ⚠️ 仅工具层, 缺组件测试 |
| 配置加载 | 无测试 | ⚠️ 低风险, 配置变更频率低 |
| CI | 基本 lint + build | ⚠️ 缺覆盖率报告门禁、并行化 |

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
│   │   │   └── testutil_test.go              # 测试辅助工具
│   │   ├── middleware/
│   │   │   └── auth_test.go                  # JWT/Token 黑名单 (217 行)
│   │   ├── service/
│   │   │   └── exchange_test.go              # 汇率服务 (423 行)
│   │   └── task/
│   │       └── tasks_test.go                 # 导入导出任务 (545 行)
│   └── infra/
│       ├── cache/
│       │   ├── memory_test.go                # 内存缓存 (182 行)
│       │   ├── tiered_test.go                # 分层缓存 (211 行)
│       │   └── cache_edge_test.go            # 缓存边缘场景 (278 行)
│       └── queue/
│           ├── queue_test.go                 # 队列标准用例 (135 行)
│           └── queue_edge_test.go            # 队列边缘场景 (126 行)
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

### Phase 5 — CI 增强（部分完成）
- [x] Go 测试 CI job (go test -race)
- [x] 前端测试 CI job (vitest)
- [ ] 覆盖率报告门禁
- [ ] go-vet 增强检查
- [ ] E2E 测试 (Docker Compose)

---

## 6. 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| Handler 测试依赖真实 PG | CI 不稳定 | 添加 `-short` flag 跳过；用 Go test DB container |
| 浮点数精度 | 断言失败 | 使用 `math.Abs(got-want) < 1e-6` 近似比较 |
| 缓存测试时序 | flaky test | 使用 `time.NewTimer` 而非 `time.Sleep`；mock 时间 |
| 前端 Ant Design 版本 | 快照测试易碎 | 不用快照测试，用逻辑断言 |
