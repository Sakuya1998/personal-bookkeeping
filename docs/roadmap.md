# 产品路线图

> 文档版本: v4.0 | 最后更新: 2026-06-02

---

## 1. 版本概览

```
v1.0 (MVP) ──── v2.0 (体验+洞察) ──── v3.0 (智能) ──── v4.0 (生态)
   2026-Q1          2026-Q2             2026-Q3         2026-Q4 →
```

---

## 2. v1.0 — MVP 基础记账 (已完成)

**代号**: "Foundation"
**目标**: 打通核心记账流程, 验证产品可行性

### 核心功能
- [x] 用户注册/登录/登出 (JWT 认证)
- [x] 账本 CRUD (多账本 + 本位币 + 存档)
- [x] 分类 CRUD (收入/支出 + 树形层级)
- [x] 交易记录 CRUD (多币种 + 自动汇率折算)
- [x] 汇率管理 (手动录入 + 反向自动计算)
- [x] 仪表盘 (收入/支出/结余 + 分类排行)
- [x] 异步导入导出 (CSV/JSON, 仅后端)

### 基础设施
- [x] PostgreSQL + GORM ORM
- [x] Docker Compose 一键部署
- [x] 多级缓存 (Memory / Redis / Tiered)
- [x] 异步任务队列 (Redis Streams / Kafka)
- [x] OpenTelemetry + Prometheus 可观测性
- [x] 结构化日志 (轮转压缩)
- [x] Swagger API 文档
- [x] CI (GitHub Actions)

---

## 3. v2.0 — 体验完善与数据洞察 (已完成)

**代号**: "Insight"
**目标**: 从"能用"到"好用", 让用户看得懂自己的钱花在哪里

### 3.1 数据可视化 (Sprint 1 · Week 1-2) ✅

| 功能 | 优先级 | 用户价值 | 状态 |
|------|--------|---------|------|
| 月度收支趋势折线图 | P0 | 用户可直观看到消费趋势变化 | ✅ |
| 分类支出饼图/环形图 | P0 | 快速了解钱花在哪里 | ✅ |
| 日历视图 (每日收支) | P0 | 按日期快速定位消费 | ✅ |

**技术实现**:
- 后端: `analytics.go` — `GET /ledgers/:id/monthly-trend`, `GET /ledgers/:id/category-breakdown`, `GET /ledgers/:id/daily-transactions`
- 前端: ECharts 折线图+环形图嵌入 DashboardPage，独立 CalendarViewPage 路由 `/ledgers/:ledger_id/calendar`
- 路由级懒加载，chunk 分割 (echarts 1.1MB 按需加载)

### 3.2 数据导入导出 (Sprint 2 · Week 3-4) ✅

| 功能 | 优先级 | 用户价值 | 状态 |
|------|--------|---------|------|
| 批量删除交易 | P0 | 清理误操作或重复数据 | ✅ |
| 批量修改分类 | P1 | 重新归类多笔交易 | ✅ |
| 前端导出 (CSV/JSON) | P1 | 导出数据做备份或分析 | ✅ |
| 标签管理 (tags 查询) | P1 | 按标签筛选交易 | ✅ |

**后端变更**:
- `POST /transactions/batch-delete` — 批量删除 + 所有权验证
- `PUT /transactions/batch-update` — 批量改分类
- `GET /ledgers/:id/export?format=csv&start_date=&end_date=` — 同步导出
- `GET /ledgers/:id/tags` — 账本标签列表

**前端变更**:
- TransactionsPage: rowSelection + 批量删除按钮 + 分类修改 Modal
- 导出按钮集成到仪表盘

### 3.3 批量操作 (Sprint 2 · Week 3-4) ✅

_已合并至 3.2，功能已实现。_

### 3.4 标签管理 (Sprint 2 · Week 3-4) ✅

_已合并至 3.2，功能已实现。_

### 3.5 设置页完善 (Sprint 3 · Week 5-6) ✅

| 功能 | 优先级 | 说明 | 状态 |
|------|--------|------|------|
| 修改密码 | P1 | 原密码验证 + 新密码 bcrypt 加密 | ✅ |
| 修改邮箱 | P1 | 新邮箱格式验证 + 唯一性检查 | ✅ |
| 用户信息展示 | P1 | 用户名只读 + 邮箱显示 | ✅ |

**后端**: `PUT /auth/password` + `PUT /auth/email`，路由已注册到 protected 组
**前端**: SettingsPage 双卡片布局 — 个人信息(只读) + 密码/邮箱修改表单

### 3.6 体验优化 (Sprint 3 · Week 5-6) ✅

| 功能 | 优先级 | 说明 | 状态 |
|------|--------|------|------|
| 搜索结果高亮 | P2 | 关键词匹配部分用 `<mark>` 包裹 | ✅ |
| 前端骨架屏 | P2 | Dashboard/Transactions 页加载占位 | ✅ |

**前端**: TransactionsPage 描述列高亮搜索关键词；DashboardPage 和 TransactionsPage 加载时显示 Skeleton 组件

### 3.7 技术债务 ✅

| 事项 | 优先级 | 说明 | 状态 |
|------|--------|------|------|
| Sprint 1/2 Handler 测试覆盖 | P1 | batch-delete/update/export/tags 测试 | ✅ |
| DB 索引优化 | P1 | `idx_transactions_ledger_user_date` + `idx_transactions_user_type` | ✅ |

**测试**: handler_sprint2_test.go 344 行，含 batch-delete/batch-update/export/tags 集成测试
**索引**: database.go createIndexes() 在 AutoMigrate 后创建

---

## 4. v3.0 — 智能记账与深度洞察 (已完成)

**代号**: "Smart"
**目标**: 降低记账成本, 提升数据价值

### 4.1 周期性交易 ✅
- [x] 支持设置重复规则: 每日/每周/每月/每年
- [x] 到日期自动创建交易记录
- [x] 典型场景: 工资、房租、订阅服务
- 实现: `handler/recurring.go` + `RecurringPage` + `task/scheduler.go` (goroutine+ticker, 同天去重)
- 路由: GET/POST/PUT/DELETE `/recurring`, GET `/recurring/upcoming`

### 4.2 报表系统 ✅
- 月度/季度 PDF 报表 ✅
- 数据: 收入趋势、支出分类、环比对比 ✅
- 支持下载 PDF (FPDF 生成) ✅
- 实现: `service/report.go` + `handler/report.go` + DashboardPage 报表按钮

### 4.3 拍照记账 ✅
- 上传小票照片 ✅
- OCR 识别金额 + 日期 + 商家 (PaddleOCR) ✅
- 自动填充交易表单 ✅
- 实现: PaddleOCR Docker 部署 + `service/ocr.go` + TransactionsPage 拍照按钮

### 4.4 支出预警 ✅
- [x] 自定义分类月预算
- [x] 当支出达到阈值时预警 (100% 触发超预算标记)
- [x] 交易表单创建时显示超预算 5 秒警告
- 实现: `handler/budget.go` (Upsert + Status + CheckBudgetOverrun) + `BudgetPage` + 交易页超预算标记

### 4.5 PWA 移动适配 ✅
- Service Worker 离线支持 ✅
- 移动端布局优化 ✅
- 主屏幕添加到桌面 ✅
- 实现: vite-plugin-pwa (Workbox generateSW), 自动注册 + 更新

### 4.6 汇率自动更新 ✅
- 接入免费汇率 API (ExchangeRate-API / Frankfurter) ✅
- 每日自动拉取最新汇率 (UTC 02:00) ✅
- 实现: scheduler + queue task, 支持 exchangerate-api 和 frankfurter 两种 provider

---

## 5. v4.0 — 生态扩展 (已完成)

**代号**: "Ecosystem"
**目标**: 构建个人财务数据枢纽

### 5.1 账本共享 ✅
- [x] 角色系统: owner / admin / member
- [x] 邀请系统: 通过邮箱邀请 + 接受/拒绝邀请
- [x] 退出账本
- [x] 权限校验中间件
- 实现: `handler/member.go` + `handler/invitation.go` + `MemberPage` + 路由权限中间件

### 5.2 角色权限 ✅
- [x] 基于角色的 UI 元素隐藏/显示
- [x] owner 独有: 账本删除、成员管理
- [x] admin 权限: 邀请管理、分类/预算编辑
- [x] member: 仅记账与查看
- 实现: `useRole` hook + 前端 RouteGuard + 后端 middleware

### 5.3 国际化 (i18n) ✅
- [x] 完整中英文翻译 (zh-CN / en-US)
- [x] 语言选择器 (用户偏好持久化到 localStorage)
- [x] 前端: react-i18next 集成
- [x] 后端错误消息中英双语
- 实现: `/public/locales/{zh-CN,en-US}/translation.json` + `useTranslation()` hook

### 5.4 币种选择器增强 ✅
- [x] 支持 100+ 币种
- [x] 按地区/常用分组: 常用币种、亚洲、欧洲、美洲等
- [x] 搜索过滤 (中文/英文/符号)
- [x] 优化用户体验: 分组折叠 + 高亮匹配
- 实现: antd `Select` + 自定义分组渲染 + 防抖搜索

### 5.5 年度财务报告 ✅
- [x] 全年收支总览 (收入 vs 支出对比)
- [x] 月度趋势明细 (12 个月逐月分析)
- [x] 消费习惯分析 (分类占比 + 同比)
- [x] 储蓄率统计
- [x] 支持导出 PDF
- 实现: `handler/annual_report.go` + `AnnualReportPage` + FPDF 模板

### 5.6 标签使用统计 ✅
- [x] 标签使用频率排行
- [x] 标签关联交易数统计
- [x] 标签分类分布
- 实现: `GET /ledgers/:id/tags/stats` + `TagStatsPage`

### 5.7 软删除 (回收站) ✅
- [x] 交易记录软删除 (deleted_at 标记)
- [x] 回收站: 查看已删除记录
- [x] 恢复删除 (还原)
- [x] 永久删除 (物理清除)
- [x] 回收站自动清理 (30 天以上自动物理删除)
- 实现: `DeletedAt` GORM 软删除 + `RecycledBinPage` + 定时清理 task

### 5.8 基础设施升级 ✅

| 事项 | 说明 | 状态 |
|------|------|------|
| golang-migrate 迁移 | 替换 AutoMigrate, 显式迁移版本控制, 支持回滚 | ✅ |
| 缓存空值哨兵保护 | Null-sentinel 防缓存穿透, 空值标记 60s TTL | ✅ |
| strutil 工具包 | `NullableStr`, `Truncate`, 字符串共用工具 | ✅ |
| 配置模块测试 | `config_test.go` 6 个测试: DSN 拼接/L1 缓存时长/默认值 | ✅ |
| OTEL 测试 | `otel_test.go` 14 个测试: Init/Shutdown/Middleware/Metrics/Meter | ✅ |
| 代码审查修复 | 31 个 Code Review 问题全部关闭 (2026-05-30) | ✅ |
| 速率限制中间件 | `ratelimit_test.go` 7 个测试: 窗口/并发/突发 | ✅ |

### 5.9 测试与质量 ✅
- [x] Service 层: auth 服务测试 (auth_test.go, 6 用例)
- [x] Service 层: ledger 工具函数测试 (ledger_test.go, 13 用例: SplitTags/CSVRow/FormatAmount/stringsJoin)
- [x] Service 层: 周期性交易测试 (recurring_test.go, 15 用例)
- [x] 配置层: config 测试 (config_test.go, 6 用例)
- [x] 可观测性: OTEL 完整测试套件 (14 用例)
- [x] 中间件: 速率限制测试 (7 用例)
- [x] 代码质量: `go vet` + `govulncheck` + `staticcheck` 全绿

---

## 6. v5.0 — 未来展望

**代号**: "AI-Native"
**目标**: 推动个人财务智能化

### 6.1 银行流水自动导入
- 对接 Plaid (海外) / 网银 CSV 解析
- 自动分类匹配
- 定期自动同步

### 6.2 自然语言记账
- 输入"今天吃饭花了 35"自动记账
- AI 分类建议
- LLM 驱动的智能查询 ("上个月交通花了多少")

### 6.3 智能预算管理
- 按月/按分类设置预算
- 预算 vs 实际对比图
- 超预算推送通知

### 6.4 投资追踪
- 对接基金/股票 API
- 投资组合视图
- 盈亏统计分析

### 6.5 多端同步
- 移动端 App (React Native / Flutter)
- 桌面端 (Tauri)
- WebDAV 自托管同步

---

## 7. 版本发布节奏

| 版本 | 预计时间 | 主要交付物 | 里程碑 |
|------|---------|-----------|--------|
| v1.0 | 2026-Q1 | MVP 记账核心 | ✅ 已完成 |
| v2.0 | 2026-Q2 | 可视化 + 批量操作 + 设置页 + 日历 | ✅ 已完成 |
| v3.0 | 2026-Q3 | 周期性交易 + 支出预警 + 报表 + 拍照记账 + PWA + 汇率自动更新 | ✅ **已交付 (2026-05-27)** |
| v4.0 | 2026-Q4 | 共享账本 + i18n + 回收站 + 年度报告 + 币种增强 + 基础设施升级 | ✅ **已交付 (2026-06-02)** |
| v5.0 | 2027-Q1 | 银行导入 + 自然语言记账 + 智能预算 + 投资追踪 | 🚧 规划中 |

---

## 8. 功能优先级矩阵

```
                    高用户价值
                        │
   v2.0 图表        │  v2.0 导入导出
   v3.0 周期性交易  │  v2.0 批量操作
   v3.0 报表        │  v4.0 共享账本
         ───────────┼─────────── 高实现成本
   低实现成本         │
         v2.0 设置  │  v3.0 OCR
         v2.0 标签  │  v4.0 年度报表
         v2.0 搜索  │  v4.0 i18n
         v2.0 高亮  │  v5.0 自然语言
                        │
                    低用户价值
```

---

## 9. 关键指标追踪

| 指标 | v1.0 | v2.0 | v3.0 | v4.0 |
|------|------|------|------|------|
| 单次记账操作步骤 | 7 步 | 5 步 | 3 步 (OCR 拍照自动填充) | 3 步 |
| 仪表盘信息维度 | 3 数字 + 1 表格 | 3 数字 + 2 图 + 1 日历 | 3 图 + 报表下载 + 超预算预警 | + 年度报表 + 标签统计 |
| 数据导出方式 | 仅后端 API | 前端一键导出 (CSV/JSON) | + PDF 报表下载 | + 年度 PDF 报告 |
| 自动汇率 | 手动录入 | 手动录入 | 自动每日更新 (UTC 02:00) | 自动每日更新 |
| 移动端支持 | 仅响应式 | 仅响应式 | PWA 离线可用 + 桌面图标 | PWA 离线可用 + 桌面图标 |
| 周期性交易 | — | — | 每日/每周/每月/每年自动生成 | + 到期提醒增强 |
| 拍照记账 | — | — | PaddleOCR 识别金额/日期/商家 | PaddleOCR 识别金额/日期/商家 |
| 多语言支持 | — | — | — | zh-CN / en-US 完整翻译 |
| 账本共享 | — | — | — | owner/admin/member 角色 + 邀请系统 |
| 软删除 | — | — | — | 回收站: 还原/永久删除/30天自动清理 |
| 币种支持 | 10+ | 10+ | 10+ | 100+ 币种, 分组 + 搜索 |
| 数据库迁移 | AutoMigrate | AutoMigrate | AutoMigrate | golang-migrate 显式版本控制 |
| 可观测性测试 | — | — | — | 14 个 OTEL 测试 |
| 代码质量修复 | — | — | — | 31 CR 问题修复 |

---

## 10. 技术债务追踪

| 事项 | 版本 | 状态 | 备注 |
|------|------|------|------|
| E2E 测试 (Docker Compose) | 待定 | ⏳ 待规划 | — |
| 前端组件单元测试 (全量) | v4.0 | ✅ | BudgetPage + RecurringPage + 状态管理 |
| Handler 全量测试 | v2.0/v3.0 | ✅ | handler_test.go + sprint2 + sprint3 |
| 缓存空值哨兵保护 | v4.0 | ✅ | 60s TTL null-sentinel |
| golang-migrate 替换 AutoMigrate | v4.0 | ✅ | 显式迁移版本控制 |
| 速率限制中间件 | v4.0 | ✅ | 内存滑动窗口 + 并发测试 |
