# 产品概述

> 文档版本: v2.0 | 最后更新: 2026-05-26

---

## 1. 产品定位

**个人记账 (Personal Bookkeeping)** 是一款面向个人用户的轻量级多账本记账工具。

- **核心理念**: 一次录入, 自动多币种折合, 给用户最干净的财务视图
- **目标场景**: 个人日常收支记录、多币种资产管理、消费数据分析
- **差异化**: 原生多币种支持、可配置的多级缓存、异步任务处理、完整的可观测性

---

## 2. 核心功能清单

### 2.1 已实现功能 (MVP)

| 功能模块 | 功能项 | 状态 |
|---------|--------|------|
| **用户认证** | 用户注册 (用户名+邮箱+密码) | ✅ |
| | 用户登录 (用户名+密码) | ✅ |
| | JWT Token 认证 | ✅ |
| | Token 黑名单撤销 (登出) | ✅ |
| | 获取当前用户信息 | ✅ |
| | 修改密码 | ✅ |
| | 修改邮箱 | ✅ |
| **账本管理** | 创建账本 (名称/描述/本位币/图标/颜色) | ✅ |
| | 编辑账本 | ✅ |
| | 删除账本 (级联删除关联交易和分类) | ✅ |
| | 账本列表 (按排序字段) | ✅ |
| | 账本详情 | ✅ |
| | 账本存档 (is_archived 软标记) | ✅ |
| | 账本汇总 (收入/支出/结余 + 分类支出排行) | ✅ |
| **分类管理** | 创建分类 (收入/支出类型) | ✅ |
| | 编辑分类 | ✅ |
| | 删除分类 (有交易关联时禁止删除) | ✅ |
| | 分类列表 (树形结构, 支持父子层级) | ✅ |
| | 分类全局共享 (跨账本) 或绑定到特定账本 | ✅ |
| | 分类缓存 (10分钟) | ✅ |
| **交易记录** | 创建交易 (多币种, 自动折算本位币) | ✅ |
| | 编辑交易 (金额/币种变更时重新折算) | ✅ |
| | 删除交易 | ✅ |
| | 交易列表 (分页 + 排序) | ✅ |
| | 多维筛选: 类型、分类、日期范围、关键词 | ✅ |
| | 关联分类信息 (Preload) | ✅ |
| | Tags 标签 (逗号分隔) | ✅ |
| | 对账标记 (is_reconciled) | ✅ |
| | 批量删除交易 | ✅ |
| | 批量修改交易分类 | ✅ |
| **汇率管理** | 创建汇率 (同日期+币种对自动覆盖) | ✅ |
| | 删除汇率 | ✅ |
| | 汇率列表 (按日期/币种筛选) | ✅ |
| | 最新汇率查询 (DISTINCT ON 每种币对一条) | ✅ |
| | 自动反向汇率计算 (USD→CNY 无记录时用 CNY→USD 取倒数) | ✅ |
| | 汇率缓存 (1小时) | ✅ |
| **仪表盘** | 总收入/总支出/结余统计卡片 | ✅ |
| | 最近5条交易记录 | ✅ |
| | 支出分类排行 (按金额降序) | ✅ |
| | 月度收支趋势折线图 (近6个月) | ✅ |
| | 分类支出环形图 (前8大分类+其他) | ✅ |
| **日历视图** | 日历每日收支汇总 (收入/支出/笔数) | ✅ |
| | 月份切换 | ✅ |
| | 点击日期查看交易明细 | ✅ |
| **系统功能** | 健康检查接口 (含 DB Ping) | ✅ |
| | Swagger API 文档 | ✅ |
| | 多级缓存 (memory / Redis / tiered) | ✅ |
| | 异步任务队列 (Redis Streams / Kafka) | ✅ |
| | 结构化日志 (文件轮转压缩) | ✅ |
| | OpenTelemetry 链路追踪 | ✅ |
| | Prometheus 指标暴露 | ✅ |
| | Docker Compose 一键部署 | ✅ |
| | CI (GitHub Actions) | ✅ |
| | 统一错误处理 | ✅ |
| **标签管理** | 标签列表 (按账本查询所有标签) | ✅ |
| | 标签搜索/筛选 (交易列表关键词过滤) | ✅ |
| **导入导出** | 异步导出交易记录 (CSV/JSON) | ✅ (仅后端) |
| | 前端导出按钮 (CSV/JSON, 含筛选条件) | ✅ |
| | 异步导入交易记录 (CSV/JSON) | ✅ (仅后端) |
| | 前端导入按钮 (含预览) | ✅ |
| | 导入模板下载 | ✅ |
| **默认数据** | 注册时自动创建默认账本 (日常账本CNY + 投资账本USD) | ✅ |
| | 注册时自动创建默认分类 (8个支出 + 5个收入) | ✅ |

### 2.2 待实现功能

| 功能模块 | 功能项 | 优先级 | 计划版本 |
|---------|--------|--------|---------|
| **进阶功能** | 周期性交易 (订阅/工资) | P1 | v3.0 |
| | 月度/季度报表 PDF | P2 | v3.0 |
| | 支出预算与预警 | P2 | v3.0 |
| | 拍照记账 (OCR) | P2 | v3.0 |
| | 标签使用统计 | P2 | v3.0 |
| | 前端骨架屏 | P2 | v3.0 |
| | 银行流水自动导入 | P3 | v4.0 |
| | 账本共享 (家庭记账) | P3 | v4.0 |

---

## 3. 用户角色

### 3.1 角色定义

| 角色 | 描述 | 权限范围 |
|------|------|---------|
| **普通用户** | 个人记账用户 | 自己的账本/分类/交易/汇率 CRUD, 个人设置 |

当前系统为纯个人单用户模式, 暂不支持多角色或多用户共享。

### 3.2 用户画像

| 画像 | 核心需求 | 使用频率 |
|------|---------|---------|
| **海外华人** | 人民币+当地币双币种记账, 自动按汇率折算 | 每日 |
| **极简记账者** | 快速记录支出, 查看月度趋势 | 每周 |
| **自由职业者** | 多项目/多收入来源分类统计 | 每月 |
| **数字游民** | 多币种(USD/EUR/JPY)资产管理 | 不定期 |

---

## 4. 核心数据模型

### 4.1 实体关系

```
User (1) ──→ Ledger (N)         # 一个用户拥有多个账本
User (1) ──→ Category (N)       # 一个用户创建多个分类
User (1) ──→ Transaction (N)    # 一个用户有多笔交易
Ledger (1) ──→ Transaction (N)  # 一个账本有多笔交易
Ledger (1) ──→ Category (N)     # 一个账本可有多个分类 (或全局共享)
Category (1) ──→ Transaction (N) # 一个分类下有多笔交易
Category (1) ──→ Category (N)   # 分类自引用父子层级
```

### 4.2 实体字段概要

| 实体 | 关键字段 |
|------|---------|
| User | id, username, email, password_hash, is_active |
| Ledger | id, user_id, name, description, base_currency, icon, color, is_archived, sort_order |
| Category | id, user_id, ledger_id (nullable), name, type(income\|expense), icon, color, parent_id, sort_order, is_active |
| Transaction | id, ledger_id, user_id, category_id, type, amount, currency, exchange_rate, base_amount, description, transaction_date, tags, is_reconciled |
| ExchangeRate | id, from_currency, to_currency, rate, date, source |

---

## 5. 技术栈

### 5.1 后端

| 组件 | 选型 | 用途 |
|------|------|------|
| 语言 | Go 1.26 | 服务端编程 |
| Web 框架 | Gin v1.12 | HTTP 路由 + 中间件 |
| ORM | GORM v1.25 | 数据库操作 |
| 数据库 | PostgreSQL 16 | 持久化存储 |
| 认证 | golang-jwt v5 + bcrypt | JWT 签发/验证 |
| 缓存抽象 | 自定义 (Cache 接口) | memory / Redis / tiered 切换 |
| 队列抽象 | 自定义 (Queue 接口) | Redis Streams / Kafka 切换 |
| 可观测性 | OpenTelemetry + Prometheus | 链路追踪 + 指标 |
| 配置管理 | Viper | YAML + 环境变量 |
| 日志 | slog + lumberjack | 结构化日志 + 轮转 |
| 数据迁移 | golang-migrate | 数据库版本管理 |
| API 文档 | Swaggo + gin-swagger | Swagger UI |

### 5.2 前端

| 组件 | 选型 | 用途 |
|------|------|------|
| 框架 | React 19 | UI 构建 |
| 语言 | TypeScript 6 | 类型安全 |
| 构建工具 | Vite 8 | 开发 + 构建 |
| UI 组件库 | Ant Design 6 + @ant-design/icons | 界面组件 |
| 状态管理 | Zustand 5 | 全局状态 |
| 路由 | React Router 7 | 页面路由 |
| HTTP 客户端 | Axios | API 调用 |
| 日期处理 | dayjs | 日期格式化/国际化 |

---

## 6. 代码统计

### 6.1 项目规模

| 类别 | 统计项 | 数量 |
|------|--------|------|
| **后端** | Handler 源文件 | 7 个 (auth/ledger/category/transaction/exchange_rate/analytics/response) |
| | API 端点总数 | 33 个 (含 Swagger/Metrics), 31 个业务端点 |
| | 测试文件 | 10 个 (_test.go) |
| **前端** | 页面组件 | 8 个 (Login/Dashboard/Transactions/Ledgers/Categories/ExchangeRates/Settings/CalendarView) |
| | 前端路由 | 8 条 |
| | 前端测试 | 3 个 (store/api/utils) |
| **集成测试** | handler_test.go | 1,029 行 |
| | handler_sprint2_test.go | 344 行 |
