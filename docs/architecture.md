# 架构设计

> 文档版本: v1.1 | 最后更新: 2026-05-29

## 更新日志

| 版本 | 日期 | 变更 |
|------|------|------|
| v1.1 | 2026-05-29 | 后端重构 Phase 1-5：handler → service 全面迁移、Cache/Queue DI 注入、database 解耦、model/→models/ 改名、infra/middleware 拆分、auth 中间件用户缓存 |

---

## 1. 整体架构概览

```
┌────────────────────────────────────────────────────────────────────┐
│                          用户 (Browser)                            │
└──────────────────────┬─────────────────────────────────────────────┘
                       │ HTTP (REST JSON)
                       ▼
┌────────────────────────────────────────────────────────────────────┐
│  ┌────────────┐       Nginx (静态资源 + 反向代理)                    │
│  │ Frontend   │  ──►  /api/* → backend:8000                       │
│  │ :3000       │       /*     → index.html (SPA)                   │
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
│  │  CORS → RateLimit → Logger → Auth (JWT + Cache) → Handler │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                    Handler Layer                          │     │
│  │  auth → service、ledger → service、category → service     │     │
│  │  transaction → service、analytics → service、budget →     │     │
│  │  exchange_rate → service、recurring → service、report →   │     │
│  │  (所有 handler 零 DB/Cache/Queue 直调)                   │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                  Service Layer (DI 注入)                  │     │
│  │  auth.go  ledger.go  category.go  transaction.go         │     │
│  │  analytics.go  budget.go  exchange_rate.go               │     │
│  │  recurring.go  report.go  ocr.go  exchange.go            │     │
│  │  └─ 通过 s.DB / s.Cache / s.Queue 访问基础设施           │     │
│  └──────────┬───────────────────────────────────────────────┘     │
│             │                                                    │
│  ┌──────────▼───────────────────────────────────────────────┐     │
│  │                 Infrastructure Layer                      │     │
│  │  ┌──────────┐  ┌────────────┐  ┌────────────┐            │     │
│  │  │ database  │  │   cache    │  │   queue    │            │     │
│  │  │ (GORM)   │  │ Memory     │  │ Inmemory   │            │     │
│  │  │   +      │  │ Redis      │  │ Redis Str  │            │     │
│  │  │ migrate  │  │ Tiered L1+ │  │ Kafka      │            │     │
│  │  │          │  │            │  │            │            │     │
│  │  ├──────────┤  ├────────────┤  ├────────────┤            │     │
│  │  │ config   │  │  logger    │  │  otel      │            │     │
│  │  │ (Viper)  │  │ (slog)    │  │ Tracing+   │            │     │
│  │  │          │  │           │  │ Metrics    │            │     │
│  │  ├──────────┤  └────────────┘  └────────────┘            │     │
│  │  │middleware│                                            │     │
│  │  │RateLimit │                                            │     │
│  └──────────┼───────────────────────────────────────────────┘     │
└─────────────┼──────────┬─────────────┬─────────────┬──────────────┘
              │          │             │             │
     ┌────────▼──┐ ┌───▼────┐ ┌───────▼───────┐ ┌───▼────────┐
     │PostgreSQL │ │ Redis  │ │ Kafka (opt)   │ │ Prometheus │
     │  :5432    │ │ :6379  │ │   :9092       │ │ (external) │
     └───────────┘ └────────┘ └───────────────┘ └────────────┘
```

---

## 2. 分层架构

### 2.1 后端分层

```
cmd/server/main.go
  └─ 应用入口: 初始化 Config → Logger → OTEL → DB → Cache → Queue
     → 注册 Task Handler → Router → 启动 HTTP Server (优雅关闭)

internal/app/router/router.go
  └─ 路由注册: URI → Middleware → Handler 映射

internal/app/handler/
  ├─ auth.go          认证 (注册/登录/登出/修改密码/邮箱)
  ├─ ledger.go        账本 CRUD + 统计/导出/标签
  ├─ category.go      分类 CRUD (树形结构)
  ├─ transaction.go   交易 CRUD + 批量操作
  ├─ analytics.go     月度趋势/分类分布/日历数据
  ├─ budget.go        预算管理 + 支出预警
  ├─ exchange_rate.go 汇率管理
  ├─ recurring.go     周期性规则 CRUD + upcoming
  ├─ report.go        PDF 报表
  └─ response.go      统一响应结构 + 错误辅助

internal/app/middleware/auth.go
  └─ JWT 认证中间件 + Token 黑名单检查 (Cache) + 用户缓存 (5min TTL)

internal/app/service/
  ├─ service.go       DI 容器: Service{DB, Cache, Queue}
  ├─ auth.go          注册/登录/修改密码/邮箱/Token 生成+黑名单
  ├─ transaction.go   创建/更新/删除/列表/批量 (含汇率折算+预算检查)
  ├─ ledger.go        账本 CRUD + 导出 CSV/JSON + 标签/统计
  ├─ category.go      分类 CRUD + 缓存管理
  ├─ analytics.go     月度趋势/分类分布/日历数据
  ├─ budget.go        预算 CRUD + 状态查询 + 超支检查
  ├─ exchange.go      汇率查询: 缓存 → DB → 反向计算
  ├─ exchange_rate.go 汇率 CRUD + 缓存失效
  ├─ recurring.go     周期性规则 CRUD + 下次执行日期计算
  ├─ report.go        PDF 报表生成 + 统计计算
  └─ ocr.go           拍照记账: OCR 服务调用 + 金额/日期/商家提取

internal/app/task/tasks.go
  └─ 异步任务: 汇率定时同步 (scheduler) + 导出任务

internal/app/models/models.go
  └─ GORM 模型定义 (User/Ledger/Category/Transaction/...)

internal/infra/
  ├─ database/database.go    DB 连接 + GetDB() 全局访问
  ├─ cache/                  缓存抽象层
  │   ├─ cache.go            Cache 接口 + Key 辅助函数
  │   ├─ memory.go           L1 内存缓存 (TTL + FIFO 淘汰)
  │   ├─ redis.go            L2 Redis 缓存
  │   ├─ tiered.go           分层缓存 (L1+L2, write-back)
  │   └─ factory.go          缓存工厂
  ├─ queue/                  队列抽象层
  │   ├─ queue.go            Queue 接口
  │   ├─ inmemory.go         内存队列 (开发/单机用)
  │   ├─ redis_streams.go    Redis Streams 实现
  │   ├─ kafka.go            Kafka 实现
  │   └─ factory.go          队列工厂
  ├─ middleware/
  │   └─ ratelimit.go        滑动窗口限流中间件
  ├─ config/config.go        Viper 配置加载 (YAML + 环境变量)
  ├─ logger/                 slog 日志 (分级文件/轮转/gin+gorm 适配)
  ├─ otel/otel.go            OpenTelemetry 初始化 (Tracing + Metrics)
  ├─ migrate/migrate.go      golang-migrate 执行
  └─ swagger/                Swagger 生成文档
```

### 2.2 前端分层

```
src/App.tsx
  └─ 根组件: BrowserRouter + ConfigProvider + Routes + ErrorBoundary

src/pages/
  ├─ LoginPage.tsx        登录/注册
  ├─ AppLayout.tsx        主布局 + 导航 (Sider + Header)
  ├─ DashboardPage.tsx    仪表盘 (折线图 + 环形图 + 统计卡片)
  ├─ LedgersPage.tsx      账本管理
  ├─ CategoriesPage.tsx   分类管理 (树形)
  ├─ TransactionsPage.tsx 交易记录 + OCR 拍照识别
  ├─── BudgetPage.tsx     预算管理
  ├─ ExchangeRatesPage.tsx 汇率管理
  ├─ RecurringPage.tsx    周期性交易
  ├─ SettingsPage.tsx     个人设置
  └─ CalendarViewPage.tsx 日历视图

src/api/
  ├─ client.ts            Axios 实例 + Token 拦截器
  └─ types.ts             TypeScript 类型定义

src/store/appStore.ts     Zustand 全局状态
src/utils/                工具函数 (币种格式化等)
src/components/           通用组件 (ErrorBoundary 等)
```

---

## 3. 数据流

### 3.1 请求处理流程 (以创建交易为例)

```
1. Browser 发送 POST /api/v1/transactions (Authorization: Bearer ***)
         │
2. Gin Router 匹配路由
         │
3. CORS Middleware — 检查 Origin
         │
4. RateLimit Middleware — 滑动窗口限流 (IP-based)
         │
5. Auth Middleware (internal/app/middleware/auth.go)
   ├─ 解析 Authorization Header
   ├─ JWT 解析 + 验签
   ├─ 检查 Token 黑名单 (Cache)
   ├─ 查询 User: 查 Cache → Miss → 查 DB → 写入 Cache (5min TTL)
   └─ 设置 c.Set("user", &user)
         │
6. TransactionHandler.Create
   ├─ 从 Context 获取 user
   ├─ 绑定 JSON → parseAmount (支持 string 或 number)
   ├─ 调用 TransactionService.CreateTransaction
   │   ├─ 校验账本归属 (DB)
   │   ├─ 默认币种 CNY
   │   ├─ 汇率折算: GetExchangeRate(currency, base, date)
   │   │   ├─ 查 Cache (exchange:rate:{from}:{to}:{date})
   │   │   ├─ 查 DB (正向→反向)
   │   │   └─ 写入 Cache (1h TTL)
   │   ├─ 计算 base_amount
   │   ├─ GORM Create → PostgreSQL
   │   ├─ 预算检查: CheckBudgetOverrun()
   │   └─ 返回 (transaction, overBudget)
         │
7. 响应: 201 { transaction, over_budget }
```

### 3.2 缓存读取流程

```
请求 → Handler → Service
  │
  ├─ 查 Cache (Key: category_list:{user_id})
  │   ├─ Hit  → 直接返回
  │   └─ Miss → 查 DB → 写入 Cache (10min TTL) → 返回
  │
缓存失效事件:
  ├─ Create/Update/Delete Category → invalidateCache()
  ├─ Create/Update ExchangeRate → invalidateCache()
  └─ 用户表查询 → 5min TTL 自动过期
```

### 3.3 分层缓存 (Tiered) 数据流

```
读取:
  Service.Cache.Get(key)
    → L1 Memory Cache (30s TTL, 10K items FIFO)
        ├─ Hit → 返回
        └─ Miss → L2 Redis Cache (300s TTL)
            ├─ Hit → 返回 + 写回 L1
            └─ Miss → DB 查询 → 写入 L2 → 写入 L1 → 返回
```

### 3.4 异步任务处理流程

```
用户请求 → Handler → Queue.Submit(task)
  │
  ├─ 同步返回 { task_id }
  │
  └─ Worker goroutine → 处理任务
      ├─ Inmemory: 内存 channel (默认, 开发/单机)
      ├─ Redis Streams: XREADGROUP → XACK
      └─ Kafka: Consume → Commit
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
│  └─ config.yaml.example →       │
│     启动时复制为 config.yaml    │
├─────────────────────────────────┤
│  依赖服务                       │
│  ├─ PostgreSQL :5432            │
│  └─ Redis :6379 (可选)          │
└─────────────────────────────────┘
```

### 4.2 配置管理

- **配置文件**: `backend/config.yaml` (从 `config.yaml.example` 复制)
- **环境变量覆盖**: Viper 支持, `DB_HOST`, `JWT_SECRET` 等
- **安全**: `config.yaml` 已加入 `.gitignore`, 不提交仓库
- **示例配置**: `config.yaml.example` 包含占位符和注释
- **配置优先级**: 环境变量 > config.yaml > 默认值

---

## 5. 关键设计决策

| 决策 | 选型 | 理由 |
|------|------|------|
| 数据库 | PostgreSQL 16 | ACID 事务, 丰富索引, JSON 查询 |
| ORM | GORM | Go 生态最成熟 ORM, AutoMigrate + Preload |
| Web 框架 | Gin | 性能好, 中间件丰富 |
| 状态管理 | Zustand | 轻量, TypeScript 友好 |
| UI 框架 | Ant Design 5 | 企业级, 中文支持好 |
| 缓存 | 分层 (L1+L2) | L1 内存快, L2 Redis 持久共享 |
| 队列 | 接口抽象 + Inmemory | 开发零依赖, 生产可切 Redis/Kafka |
| base_amount | 写入时计算 | 查询性能好, 不每次 join 汇率表 |
| Auth 用户查询 | 缓存 5min TTL | 避免 N+1 DB 查询, JWT 已保证身份 |
| Amount 输入 | any 类型兼容 | 前端 string/number 均接受 |
| Tags | 逗号分隔 text | 简单够用 |
