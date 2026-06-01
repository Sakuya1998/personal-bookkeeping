---
title: UI 重设计落地规格（方案 C）Phase 2：剩余页面深度重排
date: 2026-06-01
product: personal-bookkeeping
frontend: React + Vite + Ant Design + react-router
decisions:
  density: standard
  radius: modern
  primary_color: antd default
  categories_layout: tabs
---

## 目标

在 Phase 1 已建立 tokens + PageLayout 模板并重做关键三页的基础上，将剩余页面按统一骨架做“深度重排”，让每个页面都有稳定的信息架构与一致的交互/视觉语言：

- 统一页面骨架：标题区 + 工具条 + 内容区（加载/空态/列表态）
- 收敛容器层级：避免“列表裸出”和“重复容器”
- 提升复杂表单与多状态页面的可读性与可操作性
- 日历页统一 ledger 上下文（避免全局 Select 与页面内 Select 冲突）

## 非目标

- 不新增后端接口能力（若页面需要筛选但后端不支持，仅做 UI 结构与前端本地过滤，不承诺服务端筛选）
- 不更换 UI 库、不做暗色模式
- 不改变核心业务流程

## 全站结构约束

1. AppLayout 不再对任何路由做 `.ui-page` 包裹特例；所有页面统一使用 `PageLayout` 自己承担布局留白。
2. 每个页面必须具备：
   - `PageTitle(title, description?)`
   - `PageToolbar(left?, right?)`（可换行、窄屏两行可用）
   - 主内容容器（建议用 `Card` 承载 Empty/Skeleton/Table 三态切换）
3. 表格规范：
   - 数字/金额列右对齐
   - 操作列 icon button 保持一致尺寸与间距；可扩展动作优先收纳到 Dropdown
4. Message/Loading/Empty 规范沿用 Phase 1 文档。

## 页面级设计细则

### LedgersPage（账本管理）

重排目标：

- 页面顶栏改为 `PageLayout` header + toolbar
- 彻底消除“空态分支重复渲染一套 Modal/Form”的结构
- 将“选中账本”与“编辑/删除”交互边界更清晰，减少误触

具体要求：

- 头部：标题“账本管理”，右侧主按钮“新建账本”
- 内容：
  - 空态：内容 Card 内展示 Empty + 主按钮（触发同一套 Modal）
  - 非空态：内容 Card 内展示账本网格（现有 Card 网格可保留）
- 弹窗：唯一一份 Modal + Form，创建/编辑共用

### CategoriesPage（分类管理）

决策：保留 Tabs 结构（支出/收入）。

重排目标：

- 顶部结构统一为 `PageLayout`
- 主按钮明确当前 tab 的类型，降低误建

具体要求：

- header：标题“分类管理”，可选描述“维护收入/支出分类与排序”
- toolbar：
  - 左侧：类型 Tabs（支出/收入）
  - 右侧：主按钮为 Dropdown Button（例如“新建支出分类/新建收入分类”），默认项跟随当前 tab
- content：内容 Card 内承载 Table（支出/收入分别一个 Table，但容器与空态/加载态一致）
- 操作列：编辑/删除保持一致按钮样式；为未来扩展动作预留 Dropdown

### ExchangeRatesPage（汇率管理）

重排目标：

- 顶部加入标准工具条（币种对/日期范围/source），即使后端不支持也先形成结构
- Table 数字列更易读（右对齐、固定精度仍保留）

具体要求：

- header：标题“汇率管理”，右侧“新增汇率”
- toolbar：from/to 选择、日期范围、source（如果没有对应字段则隐藏）
- content：内容 Card 内统一加载/空态/表格

### RecurringPage（周期规则）

重排目标：

- 把“提示 Card + 主按钮”的头部区域标准化（PageHeader/Toolbar）
- 内容区用 Card 统一承载 Skeleton/Empty/Table 三态
- 弹窗表单分组降低认知负担

具体要求：

- header：标题“周期规则”，描述“自动生成订阅/租金等周期性记录”
- toolbar：右侧主按钮“新增规则”
- content：内容 Card
  - loading：Skeleton
  - empty：Empty + 主按钮
  - normal：Table
- modal：表单按“基础信息 / 规则 / 高级”分组（Divider 或 Collapse）

### BudgetPage（预算管理）

重排目标：

- 月份选择属于全页筛选，应提升为标准 toolbar
- 内容区左右两块 Card 的间距与标题层级统一
- 执行状态列表更规整（避免“自绘列表”样式漂移）

具体要求：

- header：标题“预算管理”，描述“设置分类预算并跟踪执行情况”
- toolbar：左侧 MonthPicker；右侧“新增预算”
- content：保持两列布局，但使用统一 PageLayout 容器与 Card 间距
- 执行状态：改为 List/Table 风格（行高一致、进度条对齐）

### SettingsPage（设置）

重排目标：

- 加入页面级标题区与分组结构
- 将“修改邮箱/修改密码”分组更清晰（两张 Card 或 Tabs）

具体要求：

- header：标题“设置”，描述“账号信息与安全设置”
- content：
  - Profile：个人信息 Card
  - Security：修改邮箱 Card + 修改密码 Card（推荐拆开）

### CalendarViewPage（日历视图）

重排目标：

- 接入 PageLayout + Toolbar
- 统一 ledger 上下文与 URL 语义，去除页面内重复 ledger Select
- 修正 AppLayout 的路由标题映射，避免动态路由误命中

具体要求：

- header：标题“日历视图”，描述展示当前账本/月度切换
- toolbar：月份切换（上月/下月/当前月份），必要时附加“返回账本/交易”
- ledger 上下文：
  - 优先以 URL `:ledger_id` 为准（可分享/刷新一致）
  - Header 的 ledger Select 切换时应 `navigate` 到当前页面对应 ledger 的路由（而不是仅 set store）
- 内容区：保持月历 + 当日列表，但将列表行样式收敛为可复用组件（TransactionItem）

## 验收清单（Phase 2）

- 所有页面均接入 `PageLayout`，标题区/工具条/内容区结构稳定
- 无页面出现重复 `.ui-page` 容器导致的双重 padding
- 表格页数字列右对齐，操作列一致
- CalendarViewPage：标题正确、ledger 选择与 URL 一致、无重复选择器

