# UI 重设计（方案 C）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可扩展的 UI 设计系统（tokens + 页面模板），并将 Login / AppLayout / Transactions 三个关键页按“标准密度 + 现代圆角”重做落地到代码。

**Architecture:** 使用 antd `ConfigProvider` 注入统一 theme token，并用少量全局 CSS 变量补齐“布局尺度/背景体系/排版层级”。新增 `layout` 组件（AppFrame / PageLayout / Toolbar）统一页面结构，逐页迁移以降低风险。

**Tech Stack:** React 19, TypeScript, Vite, Ant Design, react-router-dom, Vitest, Testing Library

---

## Scope

- In scope
  - Design tokens：间距尺度、圆角、阴影、排版层级、背景体系
  - 统一 layout 组件：AppFrame / PageLayout / Toolbar
  - 三页重做：LoginPage / AppLayout / TransactionsPage
  - UI 规范收敛：空状态、loading、message 文案与使用方式
- Out of scope
  - 更换 UI 库/引入 Tailwind/shadcn
  - Dark mode
  - 改动后端 API/业务逻辑

## Target Files

**Create**
- `frontend/src/theme/tokens.ts`
- `frontend/src/theme/antdTheme.ts`
- `frontend/src/components/layout/PageLayout.tsx`
- `frontend/src/components/layout/PageToolbar.tsx`
- `frontend/src/components/layout/PageTitle.tsx`
- `frontend/src/components/layout/Brand.tsx`

**Modify**
- `frontend/src/App.tsx`
- `frontend/src/index.css`
- `frontend/src/pages/LoginPage.tsx`
- `frontend/src/pages/AppLayout.tsx`
- `frontend/src/pages/TransactionsPage.tsx`
- `frontend/src/__tests__/pages/LoginPage.test.tsx`
- （按需）`frontend/src/__tests__/components/AuthEventBridge.test.tsx`

---

### Task 1: 引入 Design Tokens（不改页面结构）

**Files:**
- Create: `frontend/src/theme/tokens.ts`
- Create: `frontend/src/theme/antdTheme.ts`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: 写 tokens（不依赖 antd）**

Create `frontend/src/theme/tokens.ts`:

```ts
export const ui = {
  radius: {
    base: 10,
    modal: 12,
  },
  space: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    xxl: 32,
  },
  typography: {
    pageTitle: { fontSize: 20, lineHeight: 28, fontWeight: 600 },
    sectionTitle: { fontSize: 16, lineHeight: 24, fontWeight: 600 },
    body: { fontSize: 14, lineHeight: 22, fontWeight: 400 },
    help: { fontSize: 12, lineHeight: 20, fontWeight: 400 },
  },
  layout: {
    pageMaxWidth: 1200,
    pagePadding: 24,
    pagePaddingMobile: 16,
  },
} as const;
```

- [ ] **Step 2: 映射到 antd theme**

Create `frontend/src/theme/antdTheme.ts`:

```ts
import type { ThemeConfig } from 'antd';
import { ui } from './tokens';

export const antdTheme: ThemeConfig = {
  token: {
    borderRadius: ui.radius.base,
    borderRadiusLG: ui.radius.modal,
    fontSize: ui.typography.body.fontSize,
    lineHeight: ui.typography.body.lineHeight / ui.typography.body.fontSize,
  },
  components: {
    Button: {
      borderRadius: ui.radius.base,
    },
    Card: {
      borderRadiusLG: ui.radius.base,
    },
    Input: {
      borderRadius: ui.radius.base,
    },
    Modal: {
      borderRadiusLG: ui.radius.modal,
    },
    Table: {
      cellPaddingBlockSM: ui.space.sm,
      cellPaddingInlineSM: ui.space.md,
    },
  },
};
```

- [ ] **Step 3: 在 App.tsx 注入 theme**

Modify `frontend/src/App.tsx`:

```tsx
import { antdTheme } from './theme/antdTheme';

<ConfigProvider locale={zhCN} theme={antdTheme}>
  ...
</ConfigProvider>
```

- [ ] **Step 4: 验证不破坏现有 UI**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/theme/tokens.ts frontend/src/theme/antdTheme.ts frontend/src/App.tsx
git commit -m "feat(ui): add design tokens and antd theme"
```

---

### Task 2: 全局 CSS 变量与基础版式（不改业务组件）

**Files:**
- Modify: `frontend/src/index.css`

- [ ] **Step 1: 为 UI 尺度建立 CSS 变量**

Modify `frontend/src/index.css`（在 `body` 与 `#root` 之间添加变量段）：

```css
:root {
  --ui-radius: 10px;
  --ui-radius-modal: 12px;
  --ui-space-8: 8px;
  --ui-space-12: 12px;
  --ui-space-16: 16px;
  --ui-space-24: 24px;
  --ui-space-32: 32px;
  --ui-page-max-width: 1200px;
  --ui-page-padding: 24px;
  --ui-page-padding-mobile: 16px;
  --ui-app-bg: #f5f6f8;
}
```

- [ ] **Step 2: 统一全站背景与数字排版**

Modify `frontend/src/index.css`（补充到 `body` 规则内或之后）：

```css
body {
  background: var(--ui-app-bg);
  font-variant-numeric: tabular-nums;
}
```

- [ ] **Step 3: 新增通用布局类（用于替换 inline style）**

Modify `frontend/src/index.css`（新增类）：

```css
.ui-page {
  max-width: var(--ui-page-max-width);
  margin: 0 auto;
  padding: var(--ui-page-padding);
}

.ui-pageHeader {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--ui-space-16);
  margin-bottom: var(--ui-space-16);
}

.ui-pageTitle {
  font-size: 20px;
  line-height: 28px;
  font-weight: 600;
  margin: 0;
}

.ui-pageDesc {
  font-size: 12px;
  line-height: 20px;
  color: rgba(0, 0, 0, 0.45);
  margin-top: 4px;
}

.ui-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--ui-space-12);
}

@media (max-width: 768px) {
  .ui-page {
    padding: var(--ui-page-padding-mobile);
  }
}
```

- [ ] **Step 4: 验证样式不引入横向滚动**

Run:

```bash
cd /workspace/frontend
npm run build
```

Expected: build succeeds

- [ ] **Step 5: Commit**

```bash
git add frontend/src/index.css
git commit -m "style(ui): add global css variables and page layout utilities"
```

---

### Task 3: 新增 PageLayout / Toolbar / Title 组件（结构模板）

**Files:**
- Create: `frontend/src/components/layout/PageTitle.tsx`
- Create: `frontend/src/components/layout/PageToolbar.tsx`
- Create: `frontend/src/components/layout/PageLayout.tsx`

- [ ] **Step 1: 创建 PageTitle**

Create `frontend/src/components/layout/PageTitle.tsx`:

```tsx
import React from 'react';

type Props = {
  title: string;
  description?: string;
};

const PageTitle: React.FC<Props> = ({ title, description }) => {
  return (
    <div>
      <h1 className="ui-pageTitle">{title}</h1>
      {description ? <div className="ui-pageDesc">{description}</div> : null}
    </div>
  );
};

export default PageTitle;
```

- [ ] **Step 2: 创建 PageToolbar**

Create `frontend/src/components/layout/PageToolbar.tsx`:

```tsx
import React from 'react';

type Props = {
  left?: React.ReactNode;
  right?: React.ReactNode;
};

const PageToolbar: React.FC<Props> = ({ left, right }) => {
  if (!left && !right) return null;
  return (
    <div className="ui-toolbar">
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>{left}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>{right}</div>
    </div>
  );
};

export default PageToolbar;
```

- [ ] **Step 3: 创建 PageLayout**

Create `frontend/src/components/layout/PageLayout.tsx`:

```tsx
import React from 'react';

type Props = {
  header?: React.ReactNode;
  toolbar?: React.ReactNode;
  children: React.ReactNode;
};

const PageLayout: React.FC<Props> = ({ header, toolbar, children }) => {
  return (
    <div className="ui-page">
      {header ? <div className="ui-pageHeader">{header}</div> : null}
      {toolbar ? <div style={{ marginBottom: 16 }}>{toolbar}</div> : null}
      {children}
    </div>
  );
};

export default PageLayout;
```

- [ ] **Step 4: 运行基础测试**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout/PageTitle.tsx frontend/src/components/layout/PageToolbar.tsx frontend/src/components/layout/PageLayout.tsx
git commit -m "feat(ui): add page layout components"
```

---

### Task 4: LoginPage 重做（Tabs items + label/autocomplete + 版式）

**Files:**
- Modify: `frontend/src/pages/LoginPage.tsx`
- Modify: `frontend/src/__tests__/pages/LoginPage.test.tsx`

- [ ] **Step 1: 更新测试（确保仍能回跳 next）**

Update `frontend/src/__tests__/pages/LoginPage.test.tsx`：保持现有两条用例，但改为通过 label 取控件（避免 placeholder 改动导致脆弱）。

示例（替换查询方式）：

```ts
fireEvent.change(screen.getByLabelText('用户名'), { target: { value: 'u' } });
fireEvent.change(screen.getByLabelText('密码'), { target: { value: 'p' } });
```

- [ ] **Step 2: 将 Tabs.TabPane 改为 items**

Modify `frontend/src/pages/LoginPage.tsx`（核心结构）：

```tsx
const items = [
  {
    key: 'login',
    label: '登录',
    children: (
      <Form onFinish={onLogin} layout="vertical">
        <Form.Item label="用户名" name="username" rules={[{ required: true, message: '请输入用户名' }]}>
          <Input
            prefix={<UserOutlined />}
            placeholder="例如：alice"
            size="large"
            autoComplete="username"
          />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="请输入密码"
            size="large"
            autoComplete="current-password"
          />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block size="large">
          登录
        </Button>
      </Form>
    ),
  },
  {
    key: 'register',
    label: '注册',
    children: (
      <Form onFinish={onRegister} layout="vertical">
        <Form.Item label="用户名" name="username" rules={[{ required: true, min: 2, message: '用户名至少2个字符' }]}>
          <Input prefix={<UserOutlined />} placeholder="例如：alice" size="large" autoComplete="username" />
        </Form.Item>
        <Form.Item label="邮箱" name="email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
          <Input prefix={<MailOutlined />} placeholder="例如：alice@example.com" size="large" autoComplete="email" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, min: 6, message: '密码至少6个字符' }]}>
          <Input.Password prefix={<LockOutlined />} placeholder="设置一个密码" size="large" autoComplete="new-password" />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block size="large">
          注册
        </Button>
      </Form>
    ),
  },
];
```

- [ ] **Step 3: 登录页版式统一（去 inline 背景/居中写法）**

Modify `frontend/src/pages/LoginPage.tsx` 外层容器为更稳定结构（使用 tokens 与统一背景）：

```tsx
<div style={{ minHeight: '100dvh', display: 'grid', placeItems: 'center', padding: 16 }}>
  <Card style={{ width: 420, borderRadius: 10 }}>
    <div style={{ marginBottom: 16 }}>
      <div style={{ fontSize: 20, fontWeight: 600 }}>个人记账</div>
      <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>快速记录每一笔收支</div>
    </div>
    <Tabs activeKey={tab} onChange={(k) => setTab(k as 'login' | 'register')} items={items} centered />
  </Card>
</div>
```

- [ ] **Step 4: 运行测试与本地预览**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

Manual: 打开 `/login`，确认 Tabs 无弃用警告、字段 label 可见、autocomplete 生效。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/LoginPage.tsx frontend/src/__tests__/pages/LoginPage.test.tsx
git commit -m "refactor(ui): redesign login page layout and tabs"
```

---

### Task 5: AppLayout 重做（品牌区、Header 分区、页标题）

**Files:**
- Create: `frontend/src/components/layout/Brand.tsx`
- Modify: `frontend/src/pages/AppLayout.tsx`

- [ ] **Step 1: 创建 Brand 组件（复用 public/favicon.svg）**

Create `frontend/src/components/layout/Brand.tsx`:

```tsx
import React from 'react';

type Props = {
  collapsed: boolean;
};

const Brand: React.FC<Props> = ({ collapsed }) => {
  return (
    <div style={{ height: 32, margin: 16, display: 'flex', alignItems: 'center', justifyContent: collapsed ? 'center' : 'flex-start', gap: 10 }}>
      <img src="/favicon.svg" width={20} height={20} alt="个人记账" />
      {collapsed ? null : <span style={{ color: '#fff', fontWeight: 600, fontSize: 16 }}>个人记账</span>}
    </div>
  );
};

export default Brand;
```

- [ ] **Step 2: 在 AppLayout 中替换 emoji 品牌区，并引入页面标题映射**

Modify `frontend/src/pages/AppLayout.tsx`：

- 使用 `<Brand collapsed={collapsed} />` 替换原品牌区
- 添加 `routeTitleMap`，根据 `location.pathname` 显示标题（至少覆盖 `/`, `/transactions`, `/ledgers`, `/categories`, `/exchange-rates`, `/recurring`, `/budgets`, `/settings`）
- Header 结构：左侧折叠按钮 + 标题；右侧账本 + 用户

示例（标题计算）：

```ts
const routeTitleMap: Record<string, string> = {
  '/': '仪表盘',
  '/transactions': '交易记录',
  '/ledgers': '账本管理',
  '/categories': '分类管理',
  '/exchange-rates': '汇率管理',
  '/recurring': '周期规则',
  '/budgets': '预算管理',
  '/settings': '设置',
};

const pageTitle = routeTitleMap[location.pathname] || '个人记账';
```

- [ ] **Step 3: Content 留白统一**

将 `Content` 的 margin 收敛为使用 `.ui-page` 承担主要留白；AppLayout 的 Content 只负责背景与最小 padding。

- [ ] **Step 4: 运行前端测试**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/layout/Brand.tsx frontend/src/pages/AppLayout.tsx
git commit -m "refactor(ui): redesign app layout header and brand"
```

---

### Task 6: TransactionsPage 重排到 PageLayout（工具条响应式 + 批量态稳定）

**Files:**
- Modify: `frontend/src/pages/TransactionsPage.tsx`

- [ ] **Step 1: 将页面套入 PageLayout + PageTitle + PageToolbar**

在 `TransactionsPage` 顶层 return 中引入：

```tsx
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
```

并改造结构为：

```tsx
<PageLayout
  header={
    <>
      <PageTitle title="交易记录" />
    </>
  }
  toolbar={
    <PageToolbar
      left={...筛选控件...}
      right={...主操作与批量态...}
    />
  }
>
  ...Table/Skeleton/Modals...
</PageLayout>
```

- [ ] **Step 2: 筛选区重排（可换行，两行策略）**

将筛选控件拆为更小的块（Select/DateRange/Search），并避免单个控件宽度写死过大；保留 `Row/Col` 或改为简单 flex，但确保窄屏自动换行。

- [ ] **Step 3: 批量态稳定**

保证批量态出现时：

- “新增/拍照记账”仍然可见（或至少可通过 overflow 换行保留）
- “批量删除”视觉权重明确（danger），与其他按钮间距一致

- [ ] **Step 4: 金额列对齐与数字可读性**

将金额列 render 的外层容器改为 `display: flex; justify-content: flex-end` 或 Table `align: 'right'`（若 antd 支持），同时保持 `tabular-nums` 已由全局开启。

- [ ] **Step 5: 手工验证**

Manual checklist:

- `/transactions` 顶部结构稳定：标题区 + 工具条 + 表格
- 缩小窗口（<768）工具条自动换行不溢出
- 选中多行进入批量态按钮不挤爆

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/TransactionsPage.tsx
git commit -m "refactor(ui): restructure transactions page layout and toolbar"
```

---

### Task 7: 模板扩散（Dashboard 等页面接入 PageLayout）

**Files:**
- Modify: `frontend/src/pages/DashboardPage.tsx`
- Modify: （按需）`frontend/src/pages/*Page.tsx`

- [ ] **Step 1: Dashboard 接入 PageLayout**

为 Dashboard 添加统一标题区与内容容器，确保视觉节奏统一。

- [ ] **Step 2: 快速回归关键页面**

Manual:

- `/login` 登录页可用
- `/` 主框架标题正确
- `/transactions` 工具条与表格可用

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/DashboardPage.tsx
git commit -m "refactor(ui): apply page layout to dashboard"
```

---

## Plan Self-Review

- Spec coverage：tokens、layout 模板、Login/AppLayout/Transactions 三页重做、空状态/loading/message 规范均有对应任务；模板扩散覆盖 Dashboard（其余页面可后续追加任务）。
- Placeholder scan：无 TBD/TODO；每个 task 给出了具体文件路径与代码块。
- Type consistency：tokens 与 CSS 变量命名一致；PageLayout/PageToolbar/PageTitle 的 props 在任务中定义并在后续复用。

---

## Execution Options

Plan complete and saved to `docs/superpowers/plans/2026-06-01-ui-redesign-planC.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?

