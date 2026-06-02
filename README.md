# Personal Bookkeeping

> 轻量级多账本记账工具，主打多币种支持与简洁记账体验。

## 技术栈

| 层级 | 选型 |
|------|------|
| 前端 | React + TypeScript + Vite + Ant Design 5 + Zustand |
| 后端 | Go 1.26 + Gin + GORM |
| 数据库 | PostgreSQL |
| 缓存 | Memory / Redis / Tiered (L1+L2) |
| 队列 | Inmemory / Redis Streams / Kafka |
| 可观测性 | OpenTelemetry + Prometheus |
| 部署 | Docker Compose + Nginx |

## 快速开始

```bash
# 克隆
git clone https://github.com/Sakuya1998/personal-bookkeeping.git
cd personal-bookkeeping

# 复制配置（将占位符替换为实际值）
cp backend/config.yaml.example backend/config.yaml

# 启动
docker compose up -d

# 前端: http://localhost:3000
# API:  http://localhost:8000/api/v1
# Swagger: http://localhost:8000/swagger/index.html
```

## 功能

### v1.0 (MVP) ✅
- 用户注册/登录/登出 (JWT 认证)
- 多账本管理 (自定义本位币)
- 分类管理 (收入/支出 + 树形层级)
- 交易记录 CRUD (多币种 + 汇率自动折算)
- 汇率管理 (手动录入 + 反向计算)
- 仪表盘 (收入/支出/结余 + 分类排行)
- Docker Compose 一键部署

### v2.0 (Insight) ✅
- 数据可视化 (折线图 + 环形图)
- 日历视图 (月份切换 + 每日收支)
- 批量操作 (批量删除/修改分类)
- 数据导出 (CSV/JSON)
- 标签管理
- 设置页 (修改密码/邮箱)
- 搜索结果高亮
- Skeleton 骨架屏加载

### v3.0 (Smart) ✅
- [x] 周期性交易 (订阅/租金自动生成)
- [x] PDF 报表
- [x] 拍照记账 (OCR)
- [x] 支出预警
- [x] 汇率自动更新
- [x] PWA 移动适配

### v4.0 (Team) ✅
- [x] 共享账本 (owner/admin/member 角色 + 邀请机制)
- [x] 国际化 (zh-CN / en-US)
- [x] 币种选择器 (100+ 币种)
- [x] 标签统计
- [x] 年度报告
- [x] 软删除 (回收站)
- [x] 角色权限显隐控制
- [x] 日历视图重构

## 项目结构

```
backend/
├── cmd/server/          # 入口 (main.go)
├── internal/
│   ├── app/
│   │   ├── handler/     # HTTP handler (10 文件, 新增 member.go)
│   │   ├── middleware/   # 中间件 (auth: JWT + 缓存)
│   │   ├── models/       # GORM 模型
│   │   ├── router/       # 路由注册
│   │   ├── service/      # 业务逻辑 (12 文件, 新增 member.go)
│   │   └── task/         # 异步任务 + 调度器
│   └── infra/
│       ├── cache/        # 缓存 (memory/redis/tiered)
│       ├── config/       # Viper 配置
│       ├── database/     # DB 连接
│       ├── logger/       # slog 日志
│       ├── middleware/   # 基础设施中间件 (ratelimit)
│       ├── migrate/      # 数据库迁移
│       ├── otel/         # OpenTelemetry
│       └── queue/        # 队列 (inmemory/redis streams/kafka)
│   └── pkg/             # 通用工具包
│       └── strutil/     # 字符串处理工具
├── Dockerfile
├── config.yaml.example   # 配置模板（复制为 config.yaml 使用）
├── .gitignore            # 已排除 config.yaml
└── go.mod

frontend/
├── src/
│   ├── api/             # HTTP 客户端 + 类型定义
│   ├── pages/           # 页面组件 (12 页面, 新增 InvitePage、CalendarViewPage)
│   ├── store/           # Zustand 状态
│   ├── utils/           # 工具函数
│   └── components/      # 通用组件 (ErrorBoundary)
├── Dockerfile
├── nginx.conf
└── vite.config.ts

docker-compose.yml       # 一键部署
docs/                    # 文档
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/auth/register | 注册 |
| POST | /api/v1/auth/login | 登录 |
| POST | /api/v1/auth/logout | 登出 (撤销 Token) |
| GET | /api/v1/auth/me | 当前用户 |
| PUT | /api/v1/auth/password | 修改密码 |
| PUT | /api/v1/auth/email | 修改邮箱 |
| GET/POST | /api/v1/ledgers | 账本列表/创建 |
| GET/PUT/DELETE | /api/v1/ledgers/:id | 账本详情/更新/删除 |
| GET | /api/v1/ledgers/:id/summary | 账本统计 |
| GET | /api/v1/ledgers/:id/monthly-trend | 月度趋势 |
| GET | /api/v1/ledgers/:id/category-breakdown | 分类分布 |
| GET | /api/v1/ledgers/:id/daily-transactions | 日历数据 |
| GET | /api/v1/ledgers/:id/export | 导出 CSV/JSON |
| GET | /api/v1/ledgers/:id/tags | 标签列表 |
| GET | /api/v1/ledgers/:id/tag-stats | 标签统计 |
| GET/POST/DELETE/PUT | /api/v1/ledgers/:id/members | 成员管理 (CRUD) |
| POST | /api/v1/ledgers/:id/leave | 退出共享账本 |
| GET/POST | /api/v1/.../categories | 分类管理 |
| GET/POST | /api/v1/.../transactions | 交易记录 |
| POST | /api/v1/transactions/batch-delete | 批量删除 |
| PUT | /api/v1/transactions/batch-update | 批量改分类 |
| GET/POST | /api/v1/.../recurring | 周期性规则 CRUD + upcoming |
| GET/POST | /api/v1/.../budgets | 预算管理 + 状态查询 |
| POST | /api/v1/ocr/receipt | 拍照记账 (OCR) |
| GET | /api/v1/.../report | PDF 报表下载/预览 |
| GET/POST | /api/v1/exchange-rates | 汇率管理 |
| POST | /api/v1/exchange-rates/sync | 同步最新汇率 |
| GET | /api/v1/exchange-rates/latest | 获取最新汇率 |

完整 API 文档见 [docs/api-design.md](docs/api-design.md)

## 配置

环境变量覆盖 `config.yaml`:

| 变量 | 说明 | 默认值 |
|------|------|--------|
| DB_HOST | PostgreSQL 地址 | localhost |
| DB_PORT | PostgreSQL 端口 | 5432 |
| JWT_SECRET | JWT 密钥 | (config.yaml) |
| CACHE_TYPE | 缓存类型 | memory |
| QUEUE_TYPE | 队列类型 | inmemory |
| EXCHANGE_RATE_API_KEY | 汇率 API Key | (必填) |
| OCR_ENDPOINT | PaddleOCR API 地址 | http://localhost:9000 |

## 文档

- [架构设计](docs/architecture.md)
- [产品概述](docs/product-overview.md)
- [产品规划 v1](docs/product-plan-v1.md)
- [产品规划 v2](docs/product-plan-v2.md)
- [产品规划 v3](docs/product-plan-v3.md)
- [产品规划 v4](docs/product-plan-v4.md)
- [API 设计](docs/api-design.md)
- [路线图](docs/roadmap.md)
- [测试计划](docs/test-plan.md)
- [用户流程](docs/ux-flows.md)

## 许可

MIT
