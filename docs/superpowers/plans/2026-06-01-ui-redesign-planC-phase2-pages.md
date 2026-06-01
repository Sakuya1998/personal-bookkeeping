# UI 重设计（方案 C）Phase 2：剩余页面深度重排 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将剩余页面（Ledgers/Categories/ExchangeRates/Recurring/Budget/Settings/Calendar）按统一 PageLayout 骨架进行深度重排，收敛容器层级与交互一致性。

**Architecture:** 以 `PageLayout/PageTitle/PageToolbar` 为全站页面壳，统一标题区与工具条；将列表态（Skeleton/Empty/Table）放入一致的内容容器（Card）；消除 AppLayout 对页面容器的特例；日历页以 URL `:ledger_id` 作为 ledger 上下文权威来源并与全局 ledger Select 联动导航。

**Tech Stack:** React 19, TypeScript, Vite, Ant Design, react-router-dom, Vitest, Testing Library

---

## Preflight

- 当前工作区已有 Phase 1 改动但未提交。Phase 2 实施前建议先将 Phase 1 改动提交为一个或多个 commit，避免 Phase 2 的 diff 过大难 review。

---

## Target Files

**Modify**
- `frontend/src/pages/AppLayout.tsx`
- `frontend/src/pages/LedgersPage.tsx`
- `frontend/src/pages/CategoriesPage.tsx`
- `frontend/src/pages/ExchangeRatesPage.tsx`
- `frontend/src/pages/RecurringPage.tsx`
- `frontend/src/pages/BudgetPage.tsx`
- `frontend/src/pages/SettingsPage.tsx`
- `frontend/src/pages/CalendarViewPage.tsx`

**Create (as needed)**
- `frontend/src/components/layout/ContentCard.tsx`
- `frontend/src/components/LedgerSelect.tsx`
- `frontend/src/components/MonthNavigator.tsx`
- `frontend/src/components/TransactionItem.tsx`

**Test (as needed)**
- `frontend/src/__tests__/pages/*.test.tsx`（按页面迁移补齐最小渲染用例）

---

### Task 1: AppLayout 去特例（统一由页面自身承担 PageLayout）

**Files:**
- Modify: `frontend/src/pages/AppLayout.tsx`

- [ ] **Step 1: 移除 contentHandlesOwnLayout 与 pathname 特判**

目标：AppLayout 的 Content 区不再根据 pathname 包/不包 `.ui-page`，避免重复容器。

- [ ] **Step 2: 保持 Header/Sider 不变，确保现有 PageLayout 页面不叠加 padding**

- [ ] **Step 3: Run tests**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/AppLayout.tsx
git commit -m "refactor(ui): remove app layout page container special-cases"
```

---

### Task 2: 抽 ContentCard（统一承载 Skeleton/Empty/Table 三态）

**Files:**
- Create: `frontend/src/components/layout/ContentCard.tsx`

- [ ] **Step 1: Create ContentCard**

```tsx
import React from 'react';
import { Card } from 'antd';

type Props = {
  children: React.ReactNode;
  size?: 'default' | 'small';
};

const ContentCard: React.FC<Props> = ({ children, size = 'default' }) => {
  return (
    <Card size={size}>
      {children}
    </Card>
  );
};

export default ContentCard;
```

- [ ] **Step 2: Run tests**

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/ContentCard.tsx
git commit -m "feat(ui): add content card wrapper"
```

---

### Task 3: LedgersPage 深度重排（PageLayout + 单一 Modal）

**Files:**
- Modify: `frontend/src/pages/LedgersPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar/ContentCard**
- [ ] **Step 2: 将 Modal 提升为页面外层唯一实例，删除空态分支内的重复 Form/Modal**
- [ ] **Step 3: 空态与非空态都渲染在 ContentCard 内**
- [ ] **Step 4: Run tests + manual smoke**

```bash
cd /workspace/frontend
npm test
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/LedgersPage.tsx
git commit -m "refactor(ui): redesign ledgers page layout and modal structure"
```

---

### Task 4: CategoriesPage 深度重排（Tabs + 明确类型的新建入口）

**Files:**
- Modify: `frontend/src/pages/CategoriesPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar/ContentCard**
- [ ] **Step 2: Tabs 仍保留，但将 Tabs 放到 toolbar.left**
- [ ] **Step 3: toolbar.right 使用 Dropdown Button：默认项跟随当前 tab（新建支出/新建收入）**
- [ ] **Step 4: 表格承载在 ContentCard 内，统一空态/加载态容器**
- [ ] **Step 5: Run tests**

```bash
cd /workspace/frontend
npm test
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/CategoriesPage.tsx
git commit -m "refactor(ui): redesign categories page with toolbar tabs and typed create"
```

---

### Task 5: ExchangeRatesPage 深度重排（工具条筛选结构 + 表格一致性）

**Files:**
- Modify: `frontend/src/pages/ExchangeRatesPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar/ContentCard**
- [ ] **Step 2: toolbar.left 添加筛选 UI（from/to/dateRange/source）**
- [ ] **Step 3: 表格数值列右对齐（rate/date）**
- [ ] **Step 4: Run tests + build**

```bash
cd /workspace/frontend
npm test
npm run build
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/ExchangeRatesPage.tsx
git commit -m "refactor(ui): redesign exchange rates page toolbar and table"
```

---

### Task 6: RecurringPage 深度重排（头部提示区标准化 + 内容 Card 三态）

**Files:**
- Modify: `frontend/src/pages/RecurringPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar**
- [ ] **Step 2: 将 Skeleton/Empty/Table 三态统一放到 ContentCard 内**
- [ ] **Step 3: Modal 表单用 Divider/Collapse 分组（基础信息/规则/高级）**
- [ ] **Step 4: Run tests**

```bash
cd /workspace/frontend
npm test
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/RecurringPage.tsx
git commit -m "refactor(ui): redesign recurring page layout and form grouping"
```

---

### Task 7: BudgetPage 深度重排（月份筛选上移 + 左侧执行状态规范化）

**Files:**
- Modify: `frontend/src/pages/BudgetPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar**
- [ ] **Step 2: MonthPicker 移入 toolbar.left，新增预算放 toolbar.right**
- [ ] **Step 3: 左侧“执行状态”改为 List/Table 风格，行高与进度条对齐**
- [ ] **Step 4: Run tests**

```bash
cd /workspace/frontend
npm test
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/BudgetPage.tsx
git commit -m "refactor(ui): redesign budget page toolbar and status layout"
```

---

### Task 8: SettingsPage 深度重排（标题区 + 安全区分组）

**Files:**
- Modify: `frontend/src/pages/SettingsPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle**
- [ ] **Step 2: 将右侧“修改邮箱/修改密码”拆为两张 Card（推荐），或 Tabs（备用）**
- [ ] **Step 3: Run tests + build**

```bash
cd /workspace/frontend
npm test
npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/SettingsPage.tsx
git commit -m "refactor(ui): redesign settings page structure"
```

---

### Task 9: CalendarViewPage 深度重排（PageLayout + URL ledger 权威）

**Files:**
- Modify: `frontend/src/pages/CalendarViewPage.tsx`
- Modify: `frontend/src/pages/AppLayout.tsx`（路由标题映射支持日历）
- Create: `frontend/src/components/MonthNavigator.tsx`（如需要）

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar**
- [ ] **Step 2: 去掉页面内 Ledger Select（只保留全局 Header 的 ledger 选择）**
- [ ] **Step 3: 以 URL `ledger_id` 作为权威来源：解析参数→设置 currentLedger（仅用于 UI），切换月份只更新 month 状态**
- [ ] **Step 4: AppLayout 标题映射补齐日历（/ledgers/:id/calendar → 日历视图）**
- [ ] **Step 5: Run tests**

```bash
cd /workspace/frontend
npm test
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/CalendarViewPage.tsx frontend/src/pages/AppLayout.tsx frontend/src/components/MonthNavigator.tsx
git commit -m "refactor(ui): redesign calendar view layout and ledger context"
```

---

## Plan Self-Review

- Spec coverage：AppLayout 特例移除、所有剩余页面接入 PageLayout、Categories Tabs 决策、ContentCard 三态统一、Calendar ledger 权威来源均有对应任务。
- Placeholder scan：每个任务都有具体文件路径与明确改动目标，避免 TBD。
- Type consistency：复用 PageLayout/PageTitle/PageToolbar，新增 ContentCard 用于三态容器统一。

