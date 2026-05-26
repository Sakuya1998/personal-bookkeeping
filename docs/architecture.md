# 架构设计

> 文档版本: v1.0 | 最后更新: 2026-05-26

---

## 1. 整体架构概览

```
┌────────────────────────────────────────────────────────────────────┐
│                          用户 (Browser)                            │
└──────────────────────┬─────────────────────────────────────────────┘
                       │ HTTP (REST JSON)
                       ▼
┌────────────────────────────────────────────────────────────────────┐
│  ┌────────────┐       Nginx (静态资源 + 反向代理)                   │
│  │ Frontend   │  ──►  /api/* → backend:8000                       │
│  │ :3000      │       /*     → index.html (SPA)                    │
│  └────────────┘                                                    │
└──────────────────────┬─────────────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────────────┐
│  Backend (Go + Gin) :8000                                         │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │                      Router Layer                         │     │
│  │  /api/v1/health  /api/v1/auth/*  /api/v1/ledgers/*       │     │
│  │  /api/v1/categories/*  /api/v1/transactions/*            │     │
│  │  /api/v1/exchange-rates/*  /swagger/*  /metrics          │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │ Middleware Chain                                     │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │  CORS → Logger → Auth (JWT Bearer) → Handler              │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                    Handler Layer                          │     │
│  │  auth.go  ledger.go  category.go  transaction.go  rate.go │     │
│  │  → 请求参数校验 → 业务逻辑 → 调用 Repository → 响应      │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                  Service Layer (可选)                     │     │
│  │  services/exchange.go  — 汇率查询 + 缓存逻辑              │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                 Infrastructure Layer                      │     │
│  │                                                           │     │
│  │  ┌─────────┐  ┌──────────┐  ┌────────┐  ┌───────────┐    │     │
│  │  │ Database │  │  Cache   │  │ Queue  │  │   OTEL    │    │     │
│  │  │ (GORM)  │  │ Memory   │  │ Redis  │  │  Tracing  │    │     │
│  │  │   +     │  │ Redis    │  │Streams │  │  Metrics  │    │     │
│  │  │migrate  │  │ Tiered   │  │ Kafka  │  │           │    │     │
│  │  └────┬────┘  └────┬─────┘  └────┬───┘  └─────┬─────┘    │     │
│  └──────────┼──────────┼────────────┼─────────────┼─────────┘     │
└─────────────┼──────────┼────────────┼─────────────┼───────────────┘
              │          │            │             │
     ┌────────▼──┐ ┌───▼────┐ ┌──────▼──────┐  ┌───▼────────┐
     │PostgreSQL │ │ Redis  │ │ Kafka (opt) │  │ Prometheus │
     │  :5432    │ │ :6379  │ │   :9092     │  │  (external)│
     └───────────┘ └────────┘ └─────────────┘  └────────────┘
```

---

## 2. 分层架构

### 2.1 后端分层

```
┌────────────────────────────────────────────────────┐
│  cmd/server/main.go                                │
│  └─ 应用入口: 初始化 Config → DB → Cache → Queue    │
│     → Router → 启动 HTTP Server                    │
├────────────────────────────────────────────────────┤
│  internal/app/router/router.go                     │
│  └─ 路由注册: URI → Middleware → Handler 映射       │
├────────────────────────────────────────────────────┤
│  internal/app/handler/                             │
│  ├─ auth.go         认证相关 handler               │
│  ├─ ledger.go       账本 handler                   │
│  ├─ category.go     分类 handler (+ 缓存逻辑)       │
│  ├─ transaction.go  交易记录 handler                │
│  ├─ exchange_rate.go 汇率 handler                  │
│  └─ response.go     统一响应结构                    │
├────────────────────────────────────────────────────┤
│  internal/app/middleware/auth.go                   │
│  └─ JWT 认证中间件 + Token 黑名单检查               │
├────────────────────────────────────────────────────┤
│  internal/app/service/exchange.go                  │
│  └─ 汇率查询服务: DB 查询 → Cache 读写 → 反向计算   │
├────────────────────────────────────────────────────┤
│  internal/app/task/tasks.go                        │
│  └─ 异步任务: 导出报表 + 导入交易记录               │
├────────────────────────────────────────────────────┤
│  internal/app/model/models.go                      │
│  └─ GORM 模型定义 (User/Ledger/Category/Transaction)│
├────────────────────────────────────────────────────┤
│  internal/app/repository/database.go               │
│  └─ DB/Cache/Queue 全局访问入口 (Init/GetDB/...)   │
├────────────────────────────────────────────────────┤
│  internal/infra/                                   │
│  ├─ config/config.go       Viper 配置加载           │
│  ├─ cache/                 缓存抽象层               │
│  │   ├─ factory.go        缓存工厂                  │
│  │   ├─ memory.go         L1 内存缓存 (FIFO)       │
│  │   ├─ redis.go          L2 Redis 缓存            │
│  │   ├─ tiered.go         分层缓存 (L1+L2)         │
│  │   └─ cache.go          Cache 接口定义            │
│  ├─ queue/                队列抽象层                │
│  │   ├─ factory.go        队列工厂                  │
│  │   ├─ redis_streams.go  Redis Streams 实现        │
│  │   ├─ kafka.go          Kafka 实现               │
│  │   └─ queue.go          Queue 接口定义            │
│  ├─ migrate/migrate.go    golang-migrate 执行       │
│  ├─ migrations/           SQL 迁移文件              │
│  ├─ logger/               日志配置 (slog + Gin)     │
│  ├─ otel/otel.go          OpenTelemetry 初始化      │
│  └─ swagger/              Swagger 生成文档          │
└────────────────────────────────────────────────────┘
```

### 2.2 前端分层

```
┌────────────────────────────────────────────────────┐
│  src/App.tsx                                       │
│  └─ 根组件: BrowserRouter + ConfigProvider + Routes │
├────────────────────────────────────────────────────┤
│  src/pages/                                        │
│  ├─ LoginPage.tsx        登录/注册                 │
│  ├─ AppLayout.tsx        主布局 + 导航             │
│  ├─ DashboardPage.tsx    仪表盘                    │
│  ├─ LedgersPage.tsx      账本管理                   │
│  ├─ CategoriesPage.tsx   分类管理                   │
│  ├─ TransactionsPage.tsx 交易记录列表               │
│  ├─ ExchangeRatesPage.tsx 汇率管理                 │
│  └─ SettingsPage.tsx     个人设置                   │
├────────────────────────────────────────────────────┤
│  src/api/                                          │
│  ├─ client.ts            Axios 实例 + 拦截器        │
│  └─ types.ts             TypeScript 类型定义        │
├────────────────────────────────────────────────────┤
│  src/store/appStore.ts   Zustand 全局状态           │
├────────────────────────────────────────────────────┤
│  src/utils/                                        │
│  └─ currency.ts          币种常量和格式化工具       │
└────────────────────────────────────────────────────┘
```

---

## 3. 数据流

### 3.1 请求处理流程 (以创建交易为例)

```
1. Browser 发送 POST /api/v1/transactions (Authorization: Bearer <token>)
         │
2. Gin Router 匹配路由
         │
3. CORS Middleware — 检查 Origin
         │
4. Auth Middleware (internal/app/middleware/auth.go)
   ├─ 解析 Authorization Header
   ├─ JWT 解析 + 验签
   ├─ 检查 Token 黑名单 (Cache)
   ├─ 查询 User (DB)
   └─ 设置 c.Set("user", &user)
         │
5. TransactionHandler.Create
   ├─ 从 Context 获取 user
   ├─ 绑定 JSON 请求体 → CreateTransactionInput
   ├─ 校验账本归属 (DB: ledger.user_id = user.id)
   ├─ 校验币种, 默认 CNY
   ├─ 计算汇率: 如果交易币种 ≠ 账本位币
   │   └─ services.GetExchangeRate(from, to, date)
   │       ├─ 查 Cache (key: exchange_rate:{from}:{to}:{date})
   │       ├─ 查 DB (最近汇率 / 反向汇率)
   │       └─ 写入 Cache (1h TTL)
   ├─ 计算 base_amount = amount * exchange_rate
   ├─ 格式化 Tags ([]string → 逗号分隔字符串)
   ├─ GORM Create → PostgreSQL
   └─ Preload("Category") → 返回响应
         │
6. 响应: 201 { code: 201, message: "ok", data: { transaction... } }
```

### 3.2 缓存读取流程

```
请求 → Handler
  │
  ├─ 查 Cache (Key: category_list:{user_id})
  │   ├─ Hit  → 直接返回缓存的 JSON 数据
  │   └─ Miss → 查 DB → 序列化 → 写入 Cache (10min TTL) → 返回
  │
缓存写入/更新事件:
  └─ Create/Update/Delete Category → invalidateCategoryCache()
      └─ Cache.Delete(key)
```

### 3.3 分层缓存 (Tiered) 数据流

```
读取:
  Handler.Get(key)
    → L1 Memory Cache (30s TTL, 10K items FIFO)
        ├─ Hit → 返回
        └─ Miss → L2 Redis Cache (300s TTL)
            ├─ Hit → 返回 + 异步回写 L1
            └─ Miss → DB 查询 → 写入 L2 → 写入 L1 → 返回

写入:
  Handler.Set(key, value, ttl)
    → L2 Redis Cache.Set(ttl=300s)
    → L1 Memory Cache.Set(ttl=30s)
```

### 3.4 异步任务处理流程

```
用户请求 → Handler → Queue.Enqueue(task)
  │
  ├─ 同步返回 { task_id, status: "pending" }
  │
  └─ 后台 Worker (goroutine) ← Queue.Dequeue()
      ├─ Redis Streams: XREADGROUP → 处理 → XACK
      └─ Kafka: Consume → 处理 → Commit
          │
          ├─ ExportReportHandler
          │   ├─ 查 DB (筛选条件)
          │   ├─ 生成 CSV/JSON
          │   └─ 落盘 / 返回文件路径
          │
          └─ ImportTransactionsHandler
              ├─ 解析 CSV/JSON
              ├─ 批量写入 DB (每批 100 条)
              └─ 记录导入结果
```

---

## 4. 部署架构

### 4.1 Docker Compose (开发/生产)

```
┌─────────────────────────────────┐
│  Frontend Container             │
│  Nginx × React SPA :3000       │
│  └─ 静态资源 + API 反向代理      │
└──────────┬──────────────────────┘
           │
┌──────────▼──────────────────────┐
│  Backend Container              │
│  Go Binary :8000                │
├─────────────────────────────────┤
│  依赖服务                       │
│  ┌────────┐ ┌────────┐         │
│  │PG:5432 │ │Redis   │         │
│  │postgres│ │:6379   │         │
│  │ 16-alp │ │ 7-alp  │         │
│  └────────┘ └────────┘         │
└─────────────────────────────────┘
```

### 4.2 配置管理

- **配置文件**: `backend/config.yaml`
- **环境变量覆盖**: Viper 支持, 使用 `DB_HOST`, `JWT_SECRET` 等环境变量
- **配置热加载**: 应用启动时加载, 不支持运行时热更新
- **配置优先级**: 环境变量 > config.yaml > 默认值

---

## 5. 关键设计决策

| 决策 | 选型 | 理由 |
|------|------|------|
| 数据库 | PostgreSQL 16 | ACID 事务, 丰富的索引支持, JSON 查询 |
| ORM | GORM | Go 生态最成熟的 ORM, 支持 AutoMigrate + Preload |
| Web 框架 | Gin | 性能好, 中间件丰富, 社区活跃 |
| 状态管理 | Zustand | 轻量, TypeScript 友好, 无 Provider 嵌套 |
| UI 框架 | Ant Design 6 | 企业级组件库, 中文支持好, 开箱即用 |
| 缓存 | 分层 (L1+L2) | L1 内存快 (0 RTT), L2 Redis 持久且共享 |
| 队列 | 接口抽象 | 开发期用 Redis Streams (无额外依赖), 生产可切 Kafka |
| base_amount | 写入时计算 | 查询性能好, 不需要每次 join 汇率表 |
| 币种控制 | 不强制校验 | 用户可自定义币种, 前端 UI 仅推荐常见 12 种 |
| Tags | 逗号分隔 text | 简单够用, 不支持复杂查询 (JSONB 可后续迁移) |
