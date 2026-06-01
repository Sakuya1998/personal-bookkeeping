# UI 重设计（方案 C）Phase 2 Task7-8 执行计划（Budget/Settings）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 BudgetPage 与 SettingsPage 的深度重排：Budget 月份选择器上移到 toolbar、执行状态 Table 化、内容结构统一；Settings 接入 PageLayout 标题区并拆分安全区（修改邮箱/修改密码）。

**Architecture:** 页面统一采用 `PageLayout`（header 使用 `PageTitle`，toolbar 使用 `PageToolbar`），主要内容容器优先使用 `ContentCard` 承载加载/空态/表格三态。Budget 左侧执行状态改为 `Table`（行内展示分类、已用/预算、执行进度），右侧保留预算设置表格。

**Tech Stack:** React 19, TypeScript, Vite, Ant Design, Vitest, Testing Library

---

## Target Files

**Modify**
- `frontend/src/pages/BudgetPage.tsx`
- `frontend/src/pages/SettingsPage.tsx`
- `frontend/src/components/layout/ContentCard.tsx`（如需要支持 title/extra 等 Card Props）

---

### Task 1: ContentCard 补齐常用 Card Props（可选）

**Files:**
- Modify: `frontend/src/components/layout/ContentCard.tsx`

- [ ] **Step 1: 扩展 Props 支持 title / extra / bodyStyle / styles / className / style 等常用字段**
- [ ] **Step 2: 保持现有用法不需要修改（向后兼容）**
- [ ] **Step 3: Run tests**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

---

### Task 2: BudgetPage 重排（toolbar + 执行状态 Table 化）

**Files:**
- Modify: `frontend/src/pages/BudgetPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle/PageToolbar/ContentCard**
- [ ] **Step 2: MonthPicker 移入 toolbar.left；新增预算放 toolbar.right**
- [ ] **Step 3: “预算执行状态”改为 Table**
  - columns: 分类、已用/预算、执行（进度条+百分比）
  - pagination: false；size: small；rowKey 优先用 budget_id
- [ ] **Step 4: “预算设置”表格与“预算执行状态”统一内容容器结构（使用 ContentCard）**
- [ ] **Step 5: Run tests**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

---

### Task 3: SettingsPage 重排（PageLayout 标题区 + 安全区拆分）

**Files:**
- Modify: `frontend/src/pages/SettingsPage.tsx`

- [ ] **Step 1: 接入 PageLayout/PageTitle**
- [ ] **Step 2: “修改邮箱/修改密码”拆为两张独立 Card（同列堆叠或双列排布）**
- [ ] **Step 3: 保持现有表单逻辑与校验不变**
- [ ] **Step 4: Run tests + build**

Run:

```bash
cd /workspace/frontend
npm test
npm run build
```

Expected: PASS / build succeed

---

## Plan Self-Review

- Spec coverage：MonthPicker 上移、执行状态 Table 化、Settings 标题区接入与安全区拆分均有对应任务。
- Placeholder scan：无 TBD/TODO。
- Backward compatibility：ContentCard 可选增强，保持现有调用不破坏。
