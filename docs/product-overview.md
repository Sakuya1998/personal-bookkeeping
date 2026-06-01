# 产品概述

> 文档版本: v3.0 | 最后更新: 2026-05-29

---

## 1. 产品定位

**个人记账 (Personal Bookkeeping)** 是一款面向个人用户的轻量级多账本记账工具。

- **核心理念**: 一次录入, 自动多币种折合, 给用户最干净的财务视图
- **目标场景**: 个人日常收支记录、多币种资产管理、消费数据分析
- **差异化**: 原生多币种支持、可配置的多级缓存、异步任务处理、完整的可观测性

---

## 2. 核心功能清单

### 2.1 已实现功能

| 功能模块 | 功能项 | 状态 |
|---------|--------|------|
| **用户认证** | 用户注册 (用户名+邮箱+密码) | ✅ |
| | 用户登录 (用户名+密码) | ✅ |
| | JWT Token 认证 | ✅ |
| | Token 黑名单撤销 (Cache) | ✅ |
| | 获取当前用户信息 (用户缓存 5min TTL) | ✅ |
| | 修改密码 | ✅ |
| | 修改邮箱 | ✅ |
| **账本管理** | 创建/编辑/删除账本 (级联删除关联数据) | ✅ |
| | 账本列表 (按排序字段) + 详情 | ✅ |
| | 账本存档 (is_archived) | ✅ |
| | 账本汇总 (收入/支出/结余 + 分类排行) | ✅ |
| | 月度趋势 / 分类分布 / 日历数据 | ✅ |
| | 导出 CSV/JSON / 标签列表 | ✅ |
| **分类管理** | 创建/编辑/删除 (有交易关联时禁止) | ✅ |
| | 树形结构 (二级父子层级) | ✅ |
| | 全局共享或绑定特定账本 | ✅ |
| | 分类缓存 (10分钟, 增删改自动失效) | ✅ |
| **交易记录** | 创建/编辑/删除 (多币种 + 自动汇率折算) | ✅ |
| | 列表 (分页 + 排序 + 多维筛选) | ✅ |
| | 关联分类信息 (Preload) | ✅ |
| | Tags 标签 (逗号分隔 text) | ✅ |
| | 对账标记 (is_reconciled) | ✅ |
| | 批量删除 / 批量修改分类 | ✅ |
| **汇率管理** | 创建/删除 (同日期+币种对自动覆盖) | ✅ |
| | 列表 (按日期/币种筛选) | ✅ |
| | 最新汇率 (DISTINCT ON) | ✅ |
| | 反向汇率自动计算 | ✅ |
| | 汇率缓存 (1小时) | ✅ |
| | **自动更新** (exchangerate-api, UTC 02:00) | ✅ |
| **仪表盘** | 收入/支出/结余统计卡片 | ✅ |
| | 最近交易 + 支出分类排行 | ✅ |
| | 月度收支趋势折线图 (6个月) | ✅ |
| | 分类支出环形图 (top 8 + 其他) | ✅ |
| **日历视图** | 每日收支汇总 + 月份切换 + 交易明细 | ✅ |
| **系统功能** | 健康检查 + Swagger 文档 | ✅ |
| | 多级缓存 (memory/redis/tiered) | ✅ |
| | 异步队列 (inmemory/redis streams/kafka) | ✅ |
| | 结构化日志 (分级文件/warn/error, 轮转压缩) | ✅ |
| | OpenTelemetry (Tracing + Prometheus Metrics) | ✅ |
| | 滑动窗口限流 (IP-based RateLimiter) | ✅ |
| | CI (GitHub Actions, 3 job: backend/frontend/docker) | ✅ |
| | Docker Compose 一键部署 | ✅ |
| | config.yaml.example (不提交仓库, .gitignore 排除) | ✅ |
| **标签管理** | 标签列表 + 关键词搜索/筛选 | ✅ |
| **导入导出** | CSV/JSON 导出 (前端按钮 + 日期筛选) | ✅ |
| | 导入预览 + 批量写入 | ✅ |
| **周期性交易** | 创建/编辑/删除重复规则 (daily/weekly/monthly/yearly) | ✅ |
| | Scheduler 自动检查 + 创建交易 (同天去重) | ✅ |
| **预算管理** | 分类月预算 (Upsert/Status/Delete) | ✅ |
| | 创建交易时自动检查超支 | ✅ |
| **PDF 报表** | 月度/季度/年度 PDF (FPDF 生成) | ✅ |
| **拍照记账** | PaddleOCR 识别金额/日期/商家 | ✅ |
| | 自动填充交易表单 | ✅ |
| **PWA** | Service Worker + manifest | ✅ |
| | 离线缓存策略 + 添加到桌面 | ✅ |

### 2.2 待实现功能

| 功能模块 | 功能项 | 优先级 | 说明 |
|---------|--------|--------|------|
| **分析增强** | 年度财务报告 | P0 | 全年收支/储蓄率/分类排行 |
| | 标签使用统计 | P1 | 标签维度分析 |
| **国际化** | 多语言 (i18n) | P1 | 英文界面支持 |
| | 货币选择器增强 | P1 | 100+ 币种, 实时汇率 |
| **共享** | 账本共享 | P3 | 家庭/团队共用账本 |

---

## 3. 用户角色

### 3.1 角色定义

| 角色 | 描述 | 权限范围 |
|------|------|---------|
| **普通用户** | 个人记账用户 | 自己的账本/分类/交易/汇率/预算/规则 CRUD, 个人设置 |

当前系统为纯个人单用户模式, 暂不支持多角色或多用户共享。

### 3.2 用户画像

| 画像 | 核心需求 | 使用频率 |
|------|---------|---------|
| **海外华人** | 人民币+当地币双币种记账, 自动按汇率折算 | 每日 |
| **极简记账者** | 快速记录支出 (拍照 OCR), 查看月度趋势 | 每周 |
| **自由职业者** | 多项目/多收入来源分类统计, 预算预警 | 每月 |
| **数字游民** | 多币种 (USD/EUR/JPY) 资产管理 | 不定期 |

---

## 4. 核心数据模型

### 4.1 实体关系

```
User (1) ──→ Ledger (N)
User (1) ──→ Category (N)
User (1) ──→ Transaction (N)
User (1) ──→ RecurringRule (N)
User (1) ──→ Budget (N)

Ledger (1) ──→ Transaction (N)
Ledger (1) ──→ Category (N)      # 或全局共享 (ledger_id nullable)
Ledger (1) ──→ Budget (N)

Category (1) ──→ Transaction (N)
Category (1) ──→ Category (N)    # 自引用父子层级
Category (1) ──→ Budget (N)      # 或全局预算 (category_id nullable)
```

### 4.2 实体字段概要

| 实体 | 关键字段 |
|------|---------|
| User | id, username, email, password_hash, is_active |
| Ledger | id, user_id, name, description, base_currency, icon, color, is_archived, sort_order |
| Category | id, user_id, ledger_id (nullable), name, type(income\|expense), icon, color, parent_id, sort_order, is_active |
| Transaction | id, ledger_id, user_id, category_id, type, amount, currency, exchange_rate, base_amount, description, transaction_date, tags, is_reconciled |
| ExchangeRate | id, from_currency, to_currency, rate, date, source |
| RecurringRule | id, ledger_id, category_id, user_id, type, amount, currency, description, tags, frequency, interval, day_of_month, weekday, start_date, end_date, is_active |
| Budget | id, user_id, ledger_id, category_id (nullable), month, amount |

---

## 5. 技术栈

### 5.1 后端

| 组件 | 选型 | 用途 |
|------|------|------|
| 语言 | Go 1.26 | 服务端编程 |
| Web 框架 | Gin v1.12 | HTTP 路由 + 中间件 |
| ORM | GORM v1.25 | 数据库操作 |
| 数据库 | PostgreSQL 16 | 持久化存储 |
| 认证 | golang-jwt v5 + bcrypt | JWT 签发/验证 + 黑名单 (Cache) |
| 缓存 | 自定义 Cache 接口 | memory / Redis / tiered (L1+L2) |
| 队列 | 自定义 Queue 接口 | inmemory / Redis Streams / Kafka |
| 可观测性 | OpenTelemetry + Prometheus | 链路追踪 + 指标 |
| 配置 | Viper | YAML + 环境变量 (config.yaml.example) |
| 日志 | slog + lumberjack | 结构化日志 + 分级轮转 |
| 限流 | 滑动窗口 RateLimiter | IP 级别 |
| PDF | go-pdf/fpdf | 报表生成 |
| OCR | PaddleOCR (独立容器) | 拍照记账 |

### 5.2 前端

| 组件 | 选型 | 用途 |
|------|------|------|
| 框架 | React 19 | UI 构建 |
| 语言 | TypeScript 6 | 类型安全 |
| 构建工具 | Vite 8 | 开发 + 构建 |
| UI 组件库 | Ant Design 6 | 界面组件 |
| 图表 | ECharts (echarts-for-react) | 折线图 + 环形图 |
| 状态管理 | Zustand 5 | 全局状态 |
| 路由 | React Router 7 | 页面路由 |
| HTTP | Axios | API 调用 + 拦截器 |
| PWA | vite-plugin-pwa | Service Worker + manifest |

---

## 6. 代码统计

### 6.1 项目规模

| 类别 | 统计项 | 数量 |
|------|--------|------|
| **后端** | Handler 源文件 | 9 个 (含 analytics/budget/recurring/report/ocr) |
| | Service 源文件 | 11 个 |
| | API 端点总数 | 35+ |
| | 测试文件 | 13 个 (_test.go) |
| | 测试行数 | ~3000 行 |
| **前端** | 页面组件 | 11 个 (含 Budget/Recurring/CalendarView) |
| | 前端路由 | 10 条 |
| **集成测试** | handler_test + sprint2/3 | ~1,700 行 |
| **文档** | 文档文件 | 10 个 |
