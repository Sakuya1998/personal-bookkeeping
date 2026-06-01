---
title: UI 重设计落地规格（方案 C）
date: 2026-06-01
product: personal-bookkeeping
frontend: React + Vite + Ant Design + react-router
decisions:
  density: standard
  radius: modern
  primary_color: antd default
---

## 目标

在不引入额外 UI 库的前提下，将当前前端从“Ant 默认模板拼装”升级为“有一致性设计系统支撑的产品级 UI”，重点解决：

- 全站视觉系统缺失（间距、圆角、阴影、文字层级、背景/边框灰度不一致）
- 页面结构模板缺失（标题区、工具条、筛选区、内容区的分区不稳定）
- 高信息密度页面（交易列表）在窄屏与批量态下的可读性与可操作性不足
- 交互反馈策略不统一（loading、空状态、错误提示）

## 非目标

- 不更换组件库（继续使用 antd）
- 不引入外部图片资产作为核心依赖（尽量复用现有 `favicon.svg` / antd icons）
- 不做深色模式（后续可作为独立里程碑）
- 不改变业务功能与 API 行为（仅 UI/交互与结构重排）

## 现状入口与关键页面

- 登录/注册：`/login`
- 主框架：`AppLayout`
- 交易记录：`/transactions`（高频+高密度+多操作）
- 仪表盘：`/`（信息概览）

## 方案 C 总体策略

1. 先建立 Design Tokens（设计变量与尺度体系），并映射到 antd theme token。
2. 建立全站布局骨架与页面模板组件（AppFrame / PageLayout / Toolbar）。
3. 以“关键路径三页”率先迁移并重排：Login → AppLayout → Transactions。
4. 将模板扩散到其余页面（Dashboard 等），并统一空状态、loading、message 策略。

## 设计系统（Design Tokens）

### 密度与尺度

- 密度：标准（不追求极致紧凑，也不做大留白）
- 间距尺度（px）：4 / 8 / 12 / 16 / 24 / 32
- 圆角（现代）：全局默认 10；卡片/输入/按钮统一；弹窗可更大一档（12）
- 阴影：只保留 2 档（浮层 / 强调浮层），卡片默认不依赖重阴影

### 排版层级

- 页面标题：20/28，字重 600
- 区块标题：16/24，字重 600
- 正文：14/22
- 辅助信息：12/20，颜色更弱
- 数字列（表格金额等）：优先采用等宽数字（tabular-nums）

### 色彩语义（保持 antd 默认主色）

- Primary：沿用 antd 默认蓝
- 文字颜色：标题/正文/辅助三档（使用 antd token 或 CSS 变量映射）
- 背景体系：
  - App 背景：浅灰（登录页与内容区保持一致策略）
  - 内容容器：白色或轻微对比的浅色

## 组件与页面模板

### AppFrame（全站框架）

目标：让“侧边栏 + 顶栏 + 内容区”的层级与留白稳定、响应式可预测。

要求：

- 品牌区不使用 emoji；优先复用现有 `favicon.svg` 做图标 + 文本组合
- Header 明确分区：左侧折叠 + 当前页标题；右侧账本切换 + 用户菜单
- Content：统一外边距与最大宽度（避免页面各写各的 margin）

### PageLayout（页面结构模板）

目标：每个页面都遵循一致结构，减少“功能堆叠感”。

结构：

- PageHeader：标题（必填）+ 可选描述 + 右侧主操作区（按钮）
- PageToolbar：筛选/搜索/快捷操作；支持响应式换行；支持批量态（高度不跳变）
- PageBody：主要内容（Card/Table/Charts 等）

### Toolbar（工具条标准）

- 普通态：左侧筛选，右侧主操作
- 批量态：右侧显示“已选 n 项”与批量操作，按钮尺寸与间距保持一致
- 窄屏策略：允许自动换行到两行；优先保证主操作可见

## 页面重做细则

### 登录/注册页（LoginPage）

结构：

- 左侧（或上方）：产品标题与一句价值句（文字即可）
- 右侧（或下方）：表单卡片（Tabs 登录/注册）

要求：

- Tabs 使用 `items` 写法（消除弃用告警）
- 表单字段使用显式 label（保留 placeholder 作为示例）
- 自动填充：
  - username: `autoComplete="username"`
  - 登录密码: `autoComplete="current-password"`
  - 注册密码: `autoComplete="new-password"`
  - email: `autoComplete="email"`
- 错误提示：字段级 + 提交失败 toast（维持现有 message，但文案更可执行）

### 主布局（AppLayout）

要求：

- 侧边栏：品牌区组件化（图标 + 名称）；折叠态仍保持一致对齐
- Header：加入“当前页标题”（从路由映射，不依赖后端）
- 账本切换 Select：保持标准密度与对齐；缺省态不占位（避免布局跳动）

### 交易记录页（TransactionsPage）

结构目标：

- 顶部 PageHeader：标题“交易记录” + 右侧主按钮（新增 / 拍照记账）
- PageToolbar：筛选分区（类型/分类/日期/搜索），支持窄屏换行
- 批量态：批量按钮不挤压主按钮；展示“已选 n 项”与清晰的危险操作层级

显示目标：

- 金额列使用 tabular-nums 与右对齐（便于对比）
- Tag 密度控制（币种 tag 等尽量轻量）
- 操作列图标按钮必须有明确可点击反馈与可访问名称

## 空状态 / Loading / Message 规范

- Loading：
  - 列表首屏加载用 Skeleton
  - 局部刷新可用 Table loading，但避免同时 Skeleton+Table loading
- 空状态：
  - 列表为空显示统一空状态组件（文本 + 主操作按钮）
- Message：
  - 成功：短句
  - 失败：包含下一步（例如“请检查网络后重试”/“请重新登录”）

## 迁移顺序与里程碑（可验收）

M1：Design Tokens + AppFrame/PageLayout/Toolbar 基础落地（全站结构与风格统一起点）

- `ConfigProvider` theme token 生效
- `index.css`/CSS 变量与通用布局类生效
- 新增 layout 组件但不破坏现有路由

M2：LoginPage 重做并通过基础可用性验收

- Tabs 去弃用告警
- 表单 label 与 autocomplete 完整
- 布局、留白、圆角风格符合 tokens

M3：AppLayout 重做并通过响应式验收

- 去 emoji 品牌区
- Header 分区与页标题呈现
- Content 留白统一

M4：TransactionsPage 重排并通过高密度与窄屏验收

- 工具条在窄屏下两行/折叠仍可用
- 批量态不造成按钮跳变与主操作丢失
- 金额对齐与阅读性提升

M5：模板扩散与收敛

- Dashboard 等页面迁移到 PageLayout
- 空状态与 message 文案规范收敛

## 验收清单

- 设计系统：
  - 全站圆角风格一致（现代：10–12）
  - 页面留白尺度一致（标题区/内容区/工具条一致）
  - 文字层级清晰（标题/正文/辅助）
- 登录页：
  - Tabs 弃用告警消失
  - 表单具备 label 与正确 autocomplete
- 主布局：
  - 侧边栏品牌区专业且一致
  - Header 有稳定的标题与右侧操作对齐
- 交易页：
  - 窄屏下筛选可用、主按钮可见
  - 批量态高度稳定、危险操作明确
  - 金额对齐与可读性提升

