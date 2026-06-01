# 个人记账产品规划 v3.0

> 产品定位：面向个人用户的轻量级多账本记账工具，主打多币种支持与简洁记账体验。
> 当前阶段：v3.0 Smart 已完成，v4.0 Ecosystem 规划中

---

## 一、版本演进

```
v1.0 Foundation ── v2.0 Insight ── v3.0 Smart ── v4.0 Ecosystem
     MVP 记账          可视化+批量       智能记账          生态扩展
     2026-Q1           2026-Q2          2026-Q3           远期
```

---

## 二、v3.0 目标回顾

v3.0 Smart 的核心目标：从"记录工具"升级为"智能财务助手"。

| 目标 | 说明 | 完成度 |
|------|------|--------|
| 自动化记账 | 周期性交易、汇率自动更新、拍照 OCR 识别 | ✅ |
| 智能预警 | 预算管理、超支提醒 | ✅ |
| 数据消费 | PDF 报表、趋势分析 | ✅ |
| 移动体验 | PWA 离线可用、移动端适配 | ✅ |

---

## 三、v3.0 功能清单

### 3.1 周期性交易

**用户故事**: 作为用户，我希望设置工资/房租/订阅等周期性交易，系统自动在到期日创建交易记录，无需每月手动录入。

**后端实现**:
- 新增 `recurring_rules` 表：frequency (daily/weekly/monthly/yearly), interval, day_of_month, weekday
- 新增 `RecurringService`：Create/Update/Delete/List/Upcoming
- Scheduler goroutine：每小时检查到期规则，同天去重
- `ComputeNextRunDate` 纯函数：覆盖闰年/月末/月初等边界，含完整单元测试

**前端**:
- 独立的 RecurringPage：周期性规则列表 + 新建/编辑 Modal
- 支持按账本筛选
- Upcoming 视图：展示即将生成的交易

### 3.2 预算管理与支出预警

**用户故事**: 作为用户，我想为每个支出分类设置月度预算，当交易接近或超过预算时收到提醒。

**后端实现**:
- 新增 `budgets` 表：ledger_id, category_id (nullable = 全局预算), month, amount
- `BudgetService`：Upsert/List/Status/Delete/CheckBudgetOverrun
- CreateTransaction 时自动检查预算超支
- API 响应带 `over_budget` 标记

**前端**:
- BudgetPage：预算列表 + 进度条 (百分比) + 设置表单
- 交易创建/编辑后显示超支警告 Toast

### 3.3 PDF 报表

**用户故事**: 作为用户，我希望能将指定账本和月份的财务数据导出为 PDF 文件，方便保存和分享。

**后端实现**:
- 使用 go-pdf/fpdf 生成
- ReportService.BuildReportData：聚合收入/支出/结余/分类排行
- 支持周期选择（月度/季度/年度）
- 通过 Queue 异步生成，防止大报表阻塞请求

### 3.4 汇率自动更新

**用户故事**: 作为用户，我不希望每次手动录入汇率，系统应自动从公开 API 获取最新汇率。

**后端实现**:
- 支持 exchangerate-api (免费, 1500次/月) 和 frankfurter 两种 Provider
- 每日 UTC 02:00 自动同步 (cron scheduler)
- 启动时立即拉取一次
- 通过 `RateProvider` 接口可扩展

### 3.5 PWA 移动适配

**用户故事**: 作为用户，我希望能将应用添加到手机桌面，在离线状态下也能查看已加载的数据。

**实现**:
- vite-plugin-pwa + Service Worker
- manifest.json + 图标配置
- 离线缓存策略 (NetworkFirst / CacheFirst)
- 响应式布局：Breakpoint 适配移动端 (<=768px)

### 3.6 拍照记账 (OCR)

**用户故事**: 作为用户，我想拍一张小票照片，系统自动识别金额、日期和商家，无需手动输入。

**实现**:
- OCR 服务基于 PaddleOCR (独立容器)
- 多行文本提取：金额/日期/商家规则解析 (regex + 关键字匹配)
- 前端支持拍照/相册选择
- 识别结果自动填充交易表单 Modal

### 3.7 基础设施重构

**后端架构**:

| 变更 | 说明 |
|------|------|
| Handler → Service DI | 所有 handler 通过 Service 调用业务逻辑，零 DB/Cache/Queue 直调 |
| Cache/Queue 独立 | `cache` 包和 `queue` 包独立管理全局实例，不再依赖 `database` 包 |
| Infra 层解耦 | `database` 职责单一化为 DB 连接；新增 `infra/middleware` 存放基础设施中间件 |
| Auth 缓存 | 用户查询缓存 5 分钟 TTL，消除 N+1 DB 查询 |
| 滑动窗口限流 | 新增 `RateLimiter`，支持 IP 级别限流 |

**测试覆盖**:

| 包 | 测试内容 |
|----|----------|
| infra/cache | memory (10 cases), tiered (12 cases), edge cases (5 cases) |
| infra/queue | interface, lifecycle, factory |
| infra/config | DSN formatting, L1Duration logic (4 cases) |
| infra/middleware | RateLimiter (7 cases) |
| app/middleware | JWT Claims, AuthRequired |
| app/service | exchange, ocr, report, auth, recurring, ledger |
| app/handler | E2E HTTP integration tests (3 sprint suites) |

---

## 四、技术指标

| 指标 | v2.0 | v3.0 |
|------|------|------|
| 后端 Go 文件 | 40 | 45 |
| 后端 API 端点 | 30+ | 35+ |
| 后端单元测试 | ~2400 行 | ~3000 行 |
| 测试包 | 5 | 8 |
| 文档 | 8 文件 | 9 文件 |
| 首屏 JS | ~109KB | ~80KB (code split) |
| 缓存层级 | 3 (memory/redis/tiered) | 3 |
| 队列后端 | 2 (redis/kafka) | 3 (+inmemory) |

---

## 五、v4.0 规划 (Ecosystem)

### 5.1 核心方向

| 优先级 | 功能 | 说明 | 预估 |
|--------|------|------|------|
| P0 | 年度财务报告 | 全年收入/支出/储蓄率/分类排行 | 1 周 |
| P0 | 标签使用统计 | 标签维度分析，了解消费模式 | 0.5 周 |
| P1 | 多语言 (i18n) | 英文界面支持 | 1 周 |
| P1 | 货币选择器增强 | 支持 100+ 币种，实时汇率 | 0.5 周 |
| | P2 | 银行流水自动导入 | 支持 CSV/XLSX 格式对账单导入 | 1.5 周 |
| P3 | 账本共享 | 家庭/团队共用一个账本 | 2 周 |

### 5.2 技术债务

| 事项 | 说明 |
|------|------|
| SQL 迁移 | 当前使用 AutoMigrate，需迁移至 golang-migrate |
| 软删除 | 所有实体增加 deleted_at，支持回收站 |
| 缓存穿透保护 | Bloom filter 或空值缓存，防缓存穿透 |
| 前端 E2E 测试 | Playwright 自动化测试 |
| API 版本管理 | URL prefix /api/v2 或 Accept header |

### 5.3 远期展望

```mermaid
v4.0 Ecosystem ── v5.0 AI ── ...
  生态扩展          智能分析
  2027-Q1           2027-Q2+

v4.0 目标:
- 更完整的财务视图
- 更丰富的数据接入
- 更好的国际化
- 更强的数据安全性
```
