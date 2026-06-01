# 个人记账产品规划 v2.0

> 产品定位：面向个人用户的轻量级多账本记账工具，主打多币种支持与简洁记账体验。
> 当前阶段：v3.0 Smart 已完成，v4.0 Ecosystem 规划中

---

## 一、项目现状

### 1.1 版本演进

```
v1.0 Foundation ── v2.0 Insight ── v3.0 Smart ── v4.0 Ecosystem
     MVP 记账          可视化+批量       智能记账          生态扩展
     2026-Q1           2026-Q2          2026-Q3           远期
```

### 1.2 当前技术栈

| 层级 | 技术选型 |
|------|----------|
| 前端 | React + TypeScript + Vite + Ant Design 5 + Zustand |
| 图表 | ECharts (echarts-for-react) |
| 后端 | Go 1.26 + Gin + GORM |
| 数据库 | PostgreSQL |
| 缓存 | Memory / Redis / Tiered (L1 内存 + L2 Redis, FIFO 淘汰) |
| 队列 | Redis Streams / Kafka (可切换) |
| 认证 | JWT (bcrypt + HS256, 黑名单撤销 via Cache) |
| 可观测性 | OpenTelemetry (Trace + Prometheus Metrics) |
| 日志 | slog + lumberjack 轮转压缩 |
| 部署 | Docker Compose + Nginx |
| CI | GitHub Actions (lint + test + build) |
| 测试 | go test + vitest |

### 1.3 代码规模

| 维度 | 数据 |
|------|------|
| 后端 Go 文件 | 45 个 |
| 后端 API 端点 | 35+ |
| 前端 TSX 页面 | 11 个 |
| 后端单元测试 | ~3000 行 |
| 前端测试 | vitest + 3 个测试文件 |
| 文档 | 9 个文件, ~4000 行 |

### 1.4 领域模型

```
User (用户)
├── id, username, email, password_hash
└── Ledger (账本)
    ├── id, name, icon, base_currency, is_archived
    ├── Category (分类)
    │   ├── type: income | expense
    │   ├── name, icon
    │   └── parent_id (二级树形)
    └── Transaction (交易记录)
        ├── type: income | expense
        ├── amount, currency
        ├── exchange_rate, base_amount (自动折算)
        ├── description, tags (逗号分隔)
        ├── transaction_date
        └── category_id → Category
```

---

## 二、v2.0 功能清单（已完成）

### 2.1 数据可视化 (Sprint 1)

| 功能 | 实现方式 |
|------|---------|
| 月度收支趋势折线图 | ECharts Line, 按月聚合 SQL |
| 分类支出分布环形图 | ECharts Pie, radius ['45%', '72%'] |
| 日历视图 | 独立 CalendarViewPage, 6列 x 6行网格 |
| 交易详情弹窗 | 点击日历日期展示当日交易列表 |

**后端接口**：
- `GET /ledgers/:id/monthly-trend?months=N`
- `GET /ledgers/:id/category-breakdown?start_date=&end_date=&type=`
- `GET /ledgers/:id/daily-transactions?year=&month=`

### 2.2 批量操作 (Sprint 2)

| 功能 | 实现方式 |
|------|---------|
| 批量删除交易 | antd Table rowSelection + Modal.confirm |
| 批量修改分类 | 选中后弹出分类选择 Modal |

**后端接口**：
- `POST /transactions/batch-delete` → `{ids: [...]}`
- `PUT /transactions/batch-update` → `{ids: [...], category_id: "..."}`

**安全设计**：每次批量操作都验证所有 transaction 属于当前用户，防止越权。

### 2.3 导出 (Sprint 2)

| 功能 | 实现方式 |
|------|---------|
| CSV 导出 | 同步流式响应, Content-Disposition 附件 |
| JSON 导出 | 同数据结构, 格式切换 |
| 日期筛选 | start_date / end_date 查询参数 |
| 前端入口 | 仪表盘导出按钮 + 格式/日期选择弹窗 |

**后端接口**：`GET /ledgers/:id/export?format=csv&start_date=&end_date=`

### 2.4 标签管理 (Sprint 2)

| 功能 | 实现方式 |
|------|---------|
| 标签录入 | 创建/编辑交易时以数组传入 |
| 标签查询 | 返回该账本所有使用过的标签列表 |
| 存储 | text 列, 逗号分隔 (设计权衡: 简单够用) |

### 2.5 设置页 (Sprint 3)

| 功能 | 后端 | 前端 |
|------|------|------|
| 用户信息展示 | — | 用户名只读 + 邮箱显示 |
| 修改密码 | `PUT /auth/password` (验证旧密码 + bcrypt 加密) | 表单 (旧/新/确认) |
| 修改邮箱 | `PUT /auth/email` (格式验证 + 唯一性检查) | 表单 (邮箱输入) |

### 2.6 体验优化 (Sprint 3)

| 功能 | 说明 |
|------|------|
| 搜索结果高亮 | TransactionsPage 关键词用 `<mark>` 包裹 |
| Skeleton 骨架屏 | DashboardPage + TransactionsPage 加载占位 |
| 路由级懒加载 | DashboardPage / CalendarViewPage 按需加载 |
| Chunk 分割 | echarts / antd / vendor 分离, 首屏仅 ~109KB |

### 2.7 技术债务 (Sprint 3)

| 事项 | 说明 |
|------|------|
| DB 索引 | `idx_transactions_ledger_user_date` + `idx_transactions_user_type` |
| 集成测试 | batch-delete / batch-update / export / tags 共 344 行 |
| go vet | ✅ |
| npm run build | ✅ |
| eslint | ✅ |

---

## 三、架构设计

### 3.1 系统架构

```
                     ┌──────────────┐
                     │  Nginx (:80) │
                     │  (prod)      │
                     └──────┬───────┘
                            │
               ┌────────────┴────────────┐
               │                         │
      ┌────────┴────────┐     ┌─────────┴─────────┐
      │ Frontend (:3000) │     │  Go API (:8000)   │
      │ Vite Dev Server  │     │  Gin + GORM       │
      │ (dev)            │     │  + OpenTelemetry  │
      └─────────────────┘     └─────────┬─────────┘
                                        │
                         ┌──────────────┼──────────────┐
                         │              │              │
                  ┌──────┴──────┐ ┌─────┴─────┐ ┌─────┴─────┐
                  │ PostgreSQL  │ │  Redis    │ │ Prometheus│
                  │    :5432    │ │  :6379    │ │  :9090    │
                  └─────────────┘ └───────────┘ └───────────┘
```

### 3.2 分层架构

```
cmd/server/main.go          # 入口
├── server/                 # HTTP server 配置
│   └── router/             # 路由注册
│       └── middleware/      # Auth / CORS / OTEL
│           └── handler/    # HTTP handler (请求验证 + 响应)
│               └── service/    # 业务逻辑 (可选层)
│                   └── repository/  # 数据库访问
│                       └── model/  # GORM 模型
└── infra/                  # 基础设施
    ├── config/             # Viper 配置
    ├── logger/             # slog 日志
    ├── otel/               # OpenTelemetry
    ├── cache/              # 缓存抽象
    └── queue/              # 队列抽象
```

### 3.3 分层说明

| 层 | 职责 | 是否必须 |
|----|------|---------|
| handler | 参数解析、权限验证、响应组装 | ✅ |
| service | 业务逻辑编排（所有 handler → service DI） | ✅ |
| models | 数据模型定义 + GORM tag | ✅ |
| infra | 跨领域基础设施（不与业务耦合）：database/cache/queue/config/logger/otel/middleware | ✅ |

### 3.4 关键设计决策

| 决策 | 方案 | 评价 |
|------|------|------|
| 汇率折算 | 写入时计算 base_amount 并存储 | 查询免 join, 读性能好 |
| Tags 存储 | text 列逗号分隔 | 够用但不利于复杂查询 |
| 缓存层级 | memory / redis / tiered 可选 | 灵活, 默认 tiered |
| 队列后端 | inmemory / Redis Streams / Kafka 可选 | 开发零依赖, 默认 inmemory |
| JWT 黑名单 | 通过 cache 实现 | 轻量, 无需额外存储 |
| 代码风格 | Handler → Service (DI) → Infra | 可测试性好, 架构清晰 |
| 分类管理 | 全局分类, 跨账本可复用 | 减少重复创建 |
| Auth 用户查询 | 缓存 5min TTL | 消除 N+1 DB 查询 |

---

## 四、用户流程

### 4.1 新用户流程

```
注册 → 自动创建日常账本(CNY) + 投资账本(USD)
     → 自动创建 13 个默认分类
     → 跳转仪表盘
     → 选择账本
     → 新增第一笔交易
```

### 4.2 日常使用流程

```
登录 → 选择账本 → 仪表盘概览
     ├→ 查看折线图/环形图 (趋势分析)
     ├→ 添加交易 (选类型→分类→金额→币种→描述→标签)
     ├→ 批量操作 (勾选→删除/改分类)
     ├→ 日历视图 (月份切换→查看收支→点击详情)
     ├→ 数据导出 (选择格式/日期 → 下载)
     └→ 设置 (修改密码/邮箱)
```

### 4.3 交易筛选流程

```
交易列表页
├→ 按类型筛 (收入/支出)
├→ 按分类筛
├→ 按日期筛 (起止日期)
├→ 按关键词搜索 → 结果高亮
├→ 按账本 (账本选择器)
└→ 分页浏览
```

---

## 五、v3.0 已完成 (Smart)

### 5.1 已实现功能

| 优先级 | 功能 | 说明 |
|--------|------|------|
| P0 | 周期性交易 | 工资/房租/订阅自动创建交易, 每日/周/月/年频率 |
| P0 | 支出预警 | 分类月预算, 创建交易时检查是否超支 |
| P1 | PDF 报表 | 月度/季度 PDF, 含趋势图 + 分类统计 |
| P1 | 汇率自动更新 | 接入 exchangerate-api, 每日 UTC 02:00 定时拉取 |
| P2 | PWA 适配 | Service Worker + manifest + 添加到桌面 |
| P2 | 拍照记账 | OCR 识别小票, 自动填充金额/日期/商家 |

### 5.2 架构扩展 (v3.0)

```v3.0 新增模块:
┌────────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐
│ Recurring  │  │  Budget  │  │  Report  │  │ ExchangeRate │
│  Service   │  │  Service │  │  Service │  │ Auto-Update  │
└─────┬──────┘  └────┬─────┘  └────┬─────┘  └──────┬───────┘
      │              │             │               │
      └──────────────┼─────────────┼───────────────┘
                     │             │
            ┌────────┴─────────────┴────────┐
            │         Handler Layer          │
            └────────────────────────────────┘
```

### 5.3 v3.0 技术要点

**周期性交易**：
- 新增 `recurring_rules` 表 (frequency, day_of_month, amount, category_id, etc.)
- Scheduler goroutine 每小时检查到期规则, 自动创建交易 (同天去重)
- ComputeNextRunDate 纯函数, 含边界测试 (闰年/月末/月初)

**支出预警**：
- 新增 `budgets` 表 (ledger_id, category_id, month, amount)
- CreateTransaction 时自动检查 CheckBudgetOverrun
- 前端响应中带 over_budget 标记

**PDF 报表**：
- Go 后端用 go-pdf/fpdf 生成
- 异步导出任务 (queue handler)

**汇率自动更新**：
- 接入 exchangerate-api (免费, 1500次/月)
- UTC 02:00 定时同步 + 启动立即拉取
- 支持 exchangerate-api / frankfurter 两种 Provider

**PWA**：
- vite-plugin-pwa + Service Worker
- manifest.json + 离线缓存策略
- 响应式布局适配移动端

**拍照记账**：
- OCR 服务基于 PaddleOCR
- 多行文本提取: 金额/日期/商家规则解析
- 前端支持拍照/相册选择 + 自动填充表单

### 5.4 v3.0 领域模型扩展

```sql
-- 新增表
recurring_rules: id, ledger_id, category_id, user_id, type, amount,
                 currency, description, tags, frequency, interval,
                 day_of_month, weekday, start_date, end_date, is_active

budgets: id, user_id, ledger_id, category_id (nullable), month, amount
```

---

## 六、运营与部署

### 6.1 部署架构

```yaml
# docker-compose.yml 服务
services:
  postgres:     # 数据库
  redis:        # 缓存 + 队列
  backend:      # Go API
  frontend:     # Nginx 静态文件
```

### 6.2 CI 流水线

```yaml
jobs:
  backend-lint:   # go mod tidy → go vet → go build
  backend-test:   # go test -short
  frontend-lint:  # npm ci → eslint → tsc → vite build
  docker-build:   # docker build backend + frontend
```

### 6.3 推荐生产部署

- 前端: Nginx 反向代理 + 静态文件
- 后端: 多副本部署 (behind Nginx/ELB)
- 数据库: 托管 PostgreSQL (RDS/TiDB Cloud)
- 缓存: 托管 Redis (Upstash/ElastiCache)
- 监控: Prometheus + Grafana
- 日志: 结构化 JSON 输出, 接入 ELK/Loki

---

## 七、竞品分析

| 产品 | 定位 | 多币种 | 开源 | 技术栈 |
|------|------|--------|------|--------|
| Firefly III | 个人财务管理 | ✅ | ✅ | PHP/Laravel |
| Ledger | CLI 记账 | ✅ | ✅ | Haskell |
| Actual Budget | 预算管理 | ❌ | ✅ | JS/Electron |
| 本产品 | 轻量多币种记账 | ✅ | ✅ | Go/React |
| 随手记 | 移动记账 | ⚠️ | ❌ | 原生 App |
| MoneyWiz | 跨平台记账 | ✅ | ❌ | 原生 App |

**差异化优势**：
- 多币种自动折算是核心（非附属功能）
- 技术栈现代, 易于自托管和二次开发
- 不做大而全, 专注记账核心体验
- Docker 一键部署, Ops 成本低

---

## 八、关键指标

| 指标 | v2.0 现状 | v3.0 现状 |
|------|----------|----------|
| 单次记账操作步骤 | 5 步 | 3 步 (OCR/周期性) |
| 仪表盘信息维度 | 3 数字 + 2 图 + 1 日历 | 3 图 + 报表 + 预算看板 |
| 数据导出方式 | 前端一键导出 | 前端导出 + OCR 自动录入 |
| 汇率 | 手动录入 | 自动每日更新 (exchangerate-api) |
| 移动端支持 | 响应式 Web | PWA 离线可用 + 添加到桌面 |
| 测试覆盖率 | ~70% (后端 handler) | >85% (含 service/infra/middleware) |
| 后端代码架构 | handler 直调 DB | Handler → Service (DI) → Infra |
