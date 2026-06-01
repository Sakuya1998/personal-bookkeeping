# 个人记账产品规划 v1.0

> 产品定位：面向个人用户的轻量级多账本记账工具，主打多币种支持与简洁记账体验。
> 目标用户：有跨币种记账需求的个人用户、海外生活/工作人群、极简记账爱好者。

---

## 一、项目现状分析

### 1.1 技术栈概览

| 层级 | 技术选型 |
|------|----------|
| 前端框架 | React 19 + TypeScript 6 + Vite 8 |
| UI 组件库 | Ant Design 6 + @ant-design/icons |
| 状态管理 | Zustand 5 |
| 路由 | React Router 7 |
| HTTP 客户端 | Axios (带拦截器: Bearer token + 401 自动跳转) |
| 后端框架 | Go 1.26 + Gin |
| ORM | GORM + PostgreSQL |
| 认证 | JWT (bcrypt + HS256, 支持黑名单撤销) |
| 缓存 | 分层: memory / Redis / tiered (可配置) |
| 队列 | Redis Streams / Kafka (可配置) |
| 可观测性 | OpenTelemetry + Prometheus 指标 |
| 文档 | Swagger / Swaggo |
| 部署 | Docker Compose + Nginx |
| 数据迁移 | golang-migrate |

### 1.2 现有领域模型

```
User (用户)
└── Ledger (账本) — 拥有 base_currency, is_archived
    ├── Category (分类) — type=income|expense, 支持两级树形(父子)
    └── Transaction (交易记录) — 多币种, 自动折算 base_amount
        └── ExchangeRate (汇率) — 手动录入, 自动反向计算, 按日期排序取最新
```

### 1.3 已实现功能清单

| 功能模块 | 后端 API | 前端页面 |
|---------|---------|---------|
| ✅ 用户注册/登录/登出 | ✅ POST auth/register, login, logout, GET me | ✅ LoginPage (Tabs 登录/注册) |
| ✅ 账本 CRUD | ✅ GET/POST/PUT/DELETE ledgers, GET summary | ✅ LedgersPage (卡片视图) |
| ✅ 分类 CRUD | ✅ GET/POST/PUT/DELETE categories (含树形结构) | ✅ CategoriesPage (Tabs 收入/支出, 表格) |
| ✅ 交易记录 CRUD | ✅ GET/POST/PUT/DELETE transactions (分页+筛选) | ✅ TransactionsPage (表格+筛选+Modal表单) |
| ✅ 汇率管理 | ✅ GET/POST/DELETE exchange-rates, GET latest | ✅ ExchangeRatesPage (表格+新增表单) |
| ✅ 仪表盘 | ✅ GET /ledgers/:id/summary, GET monthly-trend, GET category-breakdown | ✅ DashboardPage (统计卡片+折线图+环形图) |
| ✅ 日历视图 | ✅ GET /ledgers/:id/daily-transactions | ✅ CalendarViewPage (月份切换+每日收支+交易详情) |
| ✅ 批量操作 | ✅ POST /transactions/batch-delete, PUT /transactions/batch-update | ✅ TransactionsPage (rowSelection+批量删除+分类修改Modal) |
| ✅ 导出 | ✅ GET /ledgers/:id/export?format=csv\|json | ✅ 仪表盘导出按钮 |
| ✅ 设置页 | ✅ PUT /auth/password, PUT /auth/email | ✅ SettingsPage (修改密码+修改邮箱) |
| ✅ 标签管理 | ✅ GET /ledgers/:id/tags | ✅ 标签在交易表单中录入 |
| ✅ 健康检查 | ✅ GET /health (含 DB ping) | — |
| ✅ 默认数据 | ✅ 注册时自动创建两个账本 + 13 个分类 | — |

### 1.4 基础设施层

- **缓存**: 支持 memory (FIFO evict)、Redis、tiered (L1 memory + L2 Redis)
- **队列**: 支持 Redis Streams (Consumer Group) 和 Kafka, 带重试
- **日志**: 分级日志文件 (info/warn/error), lumberjack 轮转压缩
- **可观测性**: OpenTelemetry SDK, Prometheus metrics endpoint, Gin 中间件
- **CORS**: 可配置 Origin 白名单
- **环境配置**: config.yaml + 环境变量覆盖 (Viper)

### 1.5 系统架构图

```
[Nginx/Frontend:3000] → [Go API:8000] → [PostgreSQL:5432]
                             ├→ [Redis:6379] (cache + queue streams)
                             ├→ [Kafka:9092] (optional queue)
                             └→ [Prometheus /metrics]
```

### 1.6 当前用户流程

```
初次使用:
  注册 → 自动创建【日常账本(CNY)】+【投资账本(USD)】+ 13个默认分类
       → 进入仪表盘 (需要先选账本)
       → 新增交易记录 (选类型→分类→金额→币种→日期→标签)
       → 查看统计 (收入/支出/结余 + 分类排行)

日常使用:
  登录 → 选择账本 → 增删改交易记录 / 查看仪表盘
```

### 1.7 关键设计决策评价

| 决策 | 评价 |
|------|------|
| 品类可跨账本 (ledger_id nullable) | ✅ 合理, 用户一次创建全局复用 |
| Tags 用逗号分隔字符串存 text 列 | ⚠️ 够用但不利于复杂查询 |
| 汇率按日期取最近值 | ✅ 灵活, 支持历史汇率追溯 |
| base_amount 在写入时计算并存储 | ✅ 查询时不需要 join, 性能好 |
| 删除账本时 cascade 删交易+分类 | ⚠️ 合理但无软删除 |
| 分类删除前检查是否有交易引用 | ✅ 防止数据不一致 |
| JWT 黑名单通过 cache 实现 | ✅ 轻量级 token 撤销 |
| 异步任务队列可切换实现 | ✅ 架构灵活性好 |
| GORM AutoMigrate 用于开发 | ✅ 便捷, 但生产需 migrate CLI |

---

## 二、产品定位与原则

### 2.1 核心价值主张

> **"一次录入，全币种自动折合，给你最干净的财务视图。"**

### 2.2 设计原则

1. **记账成本越低越好** — 每多一个操作步骤流失 20% 用户
2. **多币种是一等公民** — 每个交易自带币种，本位币自动折算
3. **不做大而全的理财软件** — 不做预算、不做投资组合、不做账单分期
4. **数据可进出** — 导入导出必须原生支持，避免供应商锁定
5. **移动优先** — 主要使用场景在手机上（虽然第一期是 Web）

### 2.3 用户画像

| 画像 | 核心需求 | 频率 |
|------|---------|------|
| 海外华人 | 人民币+当地币双币种记账, 自动按汇率折算 | 每日 |
| 极简记账者 | 快速记录支出, 查看月度趋势 | 每周 |
| 自由职业者 | 多项目/多收入来源分类统计 | 每月 |
| 币圈用户 | 多币种(USD/EUR/JPY)资产追踪 | 不定期 |

---

## 三、产品路线图

### Phase 1 ✅ 已完成 — MVP 记账核心
> 周期: 已完成 | 交付: 基础记账流程跑通

- 用户注册/登录
- 账本管理 (多账本 + 切换)
- 分类管理 (收入/支出 + 树形)
- 交易记录 CRUD (多币种 + 分页 + 筛选)
- 汇率管理 (手动录入 + 自动折算)
- 仪表盘 (收入/支出/结余统计)
- Docker Compose 一键部署

### Phase 2 ✅ 已完成 — 体验完善与数据洞察 (v2.0 Insight)
> 周期: 已完成 | 目标: 从"能用"到"好用"

已实现功能：
- 📊 仪表盘折线图 (月度收支趋势) + 环形图 (分类支出分布)
- 📅 日历视图 (月份切换 + 每日收支 + 交易详情)
- 🗑️ 批量删除交易 + 批量修改分类
- 📤 数据导出 (CSV/JSON)
- 🏷️ 标签管理 (录入 + 查询)
- ⚙️ 设置页 (修改密码/修改邮箱)
- 🔍 交易搜索关键词高亮
- ⏳ Skeleton 骨架屏加载

### Phase 3 ✅ 已完成 — 智能记账与洞察深化 (v3.0 Smart)
> 周期: 已完成 | 目标: 数据价值最大化

已实现功能：
- 📈 月度/季度报表 (PDF 导出)
- 🔁 周期性交易 (订阅、租金、工资自动生成)
- 📸 拍照记账 (OCR 识别小票)
- 💡 支出预警 (自定义月限额)
- 📱 PWA 移动适配
- 🏷️ 智能分类建议 (基于历史)

### Phase 4 📌 远期展望
- 📊 年度财务报告
- 🔗 银行流水自动导入 (对接 Plaid/银联)
- 👥 账本共享 (家庭记账场景)
- 🤖 自然语言记账 ("今天吃饭花了 35 块")
- 📋 预算管理 + 预算 vs 实际对比

---

## 四、Phase 2 详细 PRD

### Epic 1: 数据可视化

#### US-2.1 月度收支趋势图

**用户故事**: 作为记账用户, 我想在仪表盘上看到过去 6 个月的收入和支出趋势折线图, 以便了解我的消费变化趋势。

**验收标准**:
- [ ] 仪表盘增加 "月度趋势" 卡片, 展示最近 6 个月数据
- [ ] X 轴为月份, Y 轴为金额, 两条线: 收入(绿色) + 支出(红色)
- [ ] 鼠标悬停显示具体数值
- [ ] 无数据月份显示为 0 或断点
- [ ] 按当前账本本位币单位

**技术方案**:
- 后端: `GET /ledgers/:id/monthly-trend?months=6`
  - 返回 `[{month: "2026-01", income: 15000, expense: 8200}]`
  - SQL: `SELECT to_char(transaction_date, 'YYYY-MM') as month, ... GROUP BY month`
- 前端: Ant Design Charts / 轻量 ECharts 折线图
- 缓存: 月度数据缓存 5 分钟

#### US-2.2 分类支出饼图

**用户故事**: 作为用户, 我想在仪表盘上看到支出分类的饼图或环形图, 以便直观了解各项支出占比。

**验收标准**:
- [ ] 复用现有 expense_by_category API 数据
- [ ] 环形图显示各分类名称、占比百分比、金额
- [ ] 点击扇形区域可跳转到该分类的交易列表
- [ ] 不超过 8 类时全显示, 超过 8 类则合并为 "其他"

#### US-2.3 日历视图

**用户故事**: 作为用户, 我想在交易记录页面切换为日历视图, 在日历上看到每日收支汇总, 以便快速定位某天的消费情况。

**验收标准**:
- [ ] 交易列表页新增 "列表/日历" 视图切换按钮
- [ ] 日历模式下, 每天格子显示: 支出(红色金额) 和 收入(绿色金额)
- [ ] 点击某日展开当日交易明细 (Popover 或 drawer)
- [ ] 支持月份切换 (prev/next)
- [ ] 包含月汇总 (月收入/月支出/日均支出)

**数据方案**:
- 前端可一次拉取整月数据, 或复用现有分页接口
- 推荐: `GET /ledgers/:id/transactions?start_date=2026-01-01&end_date=2026-01-31&page_size=200`

---

### Epic 2: 数据导入导出

#### US-2.4 CSV/JSON 导出

**用户故事**: 作为用户, 我想在交易记录页面点击"导出", 将当前筛选条件下的交易记录导出为 CSV 或 JSON 文件。

**验收标准**:
- [ ] 交易列表页顶部增加 "导出" 按钮 (下拉: CSV / JSON)
- [ ] 导出时应用当前筛选条件 (日期范围、类型、分类)
- [ ] 导出文件名为 `{账本名}_交易记录_{日期范围}.csv`
- [ ] CSV 包含: 日期、类型、分类、金额、币种、本位币金额、描述、标签
- [ ] 导出 5000 条以内走同步, 超过走异步队列

**技术方案**:
- 后端已有 Task handler, 新增 `GET /ledgers/:id/export?format=csv&start_date=&end_date=`
- 小数据量 (<5000): 直接流式响应, Content-Disposition: attachment
- 大数据量: 提交到 queue, 异步生成后前端轮询下载

#### US-2.5 CSV/JSON 导入

**用户故事**: 作为用户, 我想从 CSV 文件导入交易记录到当前账本, 以便将之前在其他工具中的数据迁移过来。

**验收标准**:
- [ ] 交易列表页增加 "导入" 按钮
- [ ] 支持 CSV 和 JSON 格式上传
- [ ] 上传后先预览 5 条数据, 用户确认后执行导入
- [ ] 导入失败提示具体行号 + 错误原因
- [ ] 默认绑定到当前账本和当前用户

**CSV 格式**:
```csv
date,type,category,amount,currency,description,tags
2026-01-15,expense,餐饮,35.00,CNY,午餐,外卖
```

---

### Epic 3: 用户体验增强

#### US-2.6 批量操作

**用户故事**: 作为用户, 我想勾选多条交易记录一次性删除或修改分类, 以减少重复操作。

**验收标准**:
- [ ] 交易列表每行增加 checkbox, 表头增加全选 checkbox
- [ ] 选中后顶部出现操作栏: "已选 N 条" + [批量删除] [批量修改分类]
- [ ] 批量删除: 二次确认后执行
- [ ] 批量修改分类: 弹出分类选择器, 确认后更新
- [ ] 批量操作后刷新列表

**后端**:
- `POST /transactions/batch-delete` body: `{ids: [...]}`
- `PUT /transactions/batch-update` body: `{ids: [...], category_id: "..."}`

#### US-2.7 设置页完善

**用户故事**: 作为用户, 我想在设置页面修改密码和邮箱。

**验收标准**:
- [ ] 修改密码: 原密码 + 新密码 + 确认新密码
- [ ] 修改邮箱: 新邮箱 (需要验证格式)
- [ ] 修改后立即生效
- [ ] 用户名只读显示

**后端**:
- `PUT /auth/password` body: `{old_password, new_password}`
- `PUT /auth/email` body: `{email}`

#### US-2.8 标签管理

**用户故事**: 作为用户, 我想在交易记录页面按标签筛选, 并且能查看所有使用过的标签。

**验收标准**:
- [ ] 新增筛选条件: 标签 (多选)
- [ ] 标签下拉加载该账本内所有使用过的标签
- [ ] 标签在列表中显示为 tag 样式的标签

**后端**:
- `GET /ledgers/:id/tags` → 返回 `["午餐", "外卖", "交通"]`
- Tags 筛选加到现有 transaction list 接口

---

### Epic 4: 技术债务与质量

#### US-2.9 前端错误处理统一化

- 统一 API 错误拦截和展示 (当前大量 try-catch 重复代码)
- 全局 loading 骨架屏

#### US-2.10 单元测试覆盖

- Backend: handler + service 层测试 (现有 handler_test.go 仅 basic)
- Frontend: 关键页面组件渲染测试

#### US-2.11 搜索高亮

- 关键词搜索结果中, 匹配文本高亮显示
- 纯前端实现, 后端返回现有数据

---

## 五、Phase 2 Release Plan

### Sprint 1 (Week 1-2) — 数据可视化
- US-2.1 月度收支趋势图
- US-2.2 分类支出饼图
- US-2.3 日历视图

### Sprint 2 (Week 3-4) — 数据导入导出 + 批量操作
- US-2.4 CSV/JSON 导出
- US-2.5 CSV/JSON 导入
- US-2.6 批量操作
- US-2.8 标签管理

### Sprint 3 (Week 5-6) — 设置与质量
- US-2.7 设置页完善
- US-2.9 错误处理统一化
- US-2.10 测试覆盖
- US-2.11 搜索高亮
- Bug bash + 性能优化

---

## 六、数据库变更 (Phase 2)

### 新增 migration: 000002

```sql
-- 用户密码变更历史 (审计)
CREATE TABLE IF NOT EXISTS password_history (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- 导出任务记录
CREATE TABLE IF NOT EXISTS export_tasks (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ledger_id uuid NOT NULL REFERENCES ledgers(id) ON DELETE CASCADE,
    format varchar(10) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'pending',  -- pending | processing | done | failed
    file_path text,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

-- 月度趋势视图 (物化备用, 避免反复聚合大表)
-- 实际用查询即可, 初期不建物化视图
```

---

## 七、关键指标 (OKR)

| 指标 | 当前 | Phase 2 目标 |
|------|------|-------------|
| 单次记账操作步骤 | 7 步 (登录→选账本→新增→选类型→选分类→输金额→保存) | 5 步 (合并优化) |
| 仪表盘可用信息维度 | 3 个数字 + 表格 | 3 个数字 + 2 张图 + 日历 |
| 数据导出能力 | 仅后端 API, 无 UI | CSV/JSON 一键导出 |
| 设置页完成度 | 10% | 100% (密码/邮箱) |
| 前端测试覆盖率 | <5% | >30% |

---

## 八、风险与依赖

| 风险 | 影响 | 缓解方案 |
|------|------|---------|
| 日历视图大数据量渲染 | 大月可能卡顿 | 前端虚拟滚动 + 服务端分页, 单次不超过 500 条 |
| Import 数据格式兼容 | 用户 CSV 格式不匹配 | 提供模板下载, 导入前预览校验 |
| 汇率数据手动录入不精准 | 货币折算偏差 | 默认汇率 = 1.0, 提示用户自行维护; 后续接入免费汇率 API |
| Docker Compose 不含 Kafka | 队列不可用 | 默认用 Redis Streams, Kafka 可选 |

---

## 九、UX 设计要点 (Phase 2)

### 9.1 仪表盘新布局
```
┌─────────────────────────────────────────────┐
│  总收入 ¥XX  总支出 ¥XX  结余 ¥XX           │  ← 已有
├──────────────────┬──────────────────────────┤
│  月度趋势折线图    │  分类支出环形图           │  ← NEW
│  (6个月)          │  (top 8 + 其他)          │
├──────────────────┴──────────────────────────┤
│  最近交易记录表格                             │  ← 已有
│  查看全部 →                                  │
├─────────────────────────────────────────────┤
│  支出分类排行                                 │  ← 已有
│  🍽️ 餐饮 ¥3,200   🚗 交通 ¥500             │
└─────────────────────────────────────────────┘
```

### 9.2 交易列表视图切换
```
[列表视图 | 日历视图]  [筛选区]  [导出▼] [导入] [+新增]
┌─────────────────────────────────────────────┐
│  日  一  二  三  四  五  六                  │
│        1   2   3   4   5   6                 │
│         支出¥35  收入¥0  支出¥120            │
│  7   8   9  10  11  12  13                  │
│  ...                                        │
└─────────────────────────────────────────────┘
```

---

## 十、总结

本项目一期 (MVP) 已经完成了核心记账流程的完整搭建, 技术架构干净、可扩展。Phase 2 的核心目标是:

1. **让数据说话** — 通过图表和日历视图, 让用户直观看到自己的消费模式
2. **降低数据流转成本** — 导入导出功能让用户数据自由进出
3. **提升操作效率** — 批量操作、标签筛选、快捷键
4. **完善基础体验** — 设置页、错误提示、搜索体验

Phase 2 不新增实体模型, 全部在现有 5 个实体的基础上扩展 API 和前端能力。预计 6 周完成。
