# CalendarViewPage（Phase2 Task9）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 深度重排 CalendarViewPage 接入 PageLayout/PageTitle/PageToolbar，移除页面内账本选择器并以 URL `ledger_id` 为权威；补齐 AppLayout 标题映射显示“日历视图”；全局账本切换在日历页时导航到新账本日历路由；确保 `npm test` 与 `npm run build` 通过。

**Architecture:** CalendarViewPage 的数据拉取以路由参数 `ledger_id` 为唯一权威来源；页面仍同步 store 的 `currentLedger` 仅用于全局 Header Select 显示与共享货币/名称信息，但不作为数据权威。AppLayout 在检测到当前路由为日历视图时，将账本选择器 onChange 变为“设置 store + navigate 到 `/ledgers/:id/calendar`”。

**Tech Stack:** React 19, TypeScript, Vite, Ant Design, react-router-dom, Vitest, Testing Library

---

## File Map

**Modify**
- `frontend/src/pages/CalendarViewPage.tsx`
- `frontend/src/pages/AppLayout.tsx`

**Test (add)**
- `frontend/src/__tests__/pages/CalendarViewPage.test.tsx`

---

### Task 1: CalendarViewPage 接入 PageLayout/PageTitle/PageToolbar，移除页内 Ledger Select，URL ledger_id 为权威

**Files:**
- Modify: `frontend/src/pages/CalendarViewPage.tsx`

- [ ] **Step 1: 调整 imports（引入 PageLayout/PageTitle/PageToolbar/ContentCard，移除 Select 相关）**

- [ ] **Step 2: 以 `ledger_id` 作为数据拉取 ledgerId（不再依赖 store 的 currentLedger.id）**

关键点：
- `const urlLedgerId = ledger_id || ''`
- `const ledgerFromUrl = ledgers.find(l => l.id === urlLedgerId) || null`
- 同步 store：当 `ledgerFromUrl` 存在且 `currentLedger?.id !== urlLedgerId` 时调用 `setCurrentLedger(ledgerFromUrl)`
- 获取数据：使用 `urlLedgerId`

- [ ] **Step 3: 在 `urlLedgerId` 或 `currentMonth` 变化时清空 selectedDate/dayTxns/dailyData 并重拉取该月数据**

- [ ] **Step 4: 页面骨架切换为 PageLayout**

结构：
- header: `<PageTitle title="日历视图" description={ledgerFromUrl ? `当前账本：${ledgerFromUrl.name}` : undefined} />`
- toolbar: `<PageToolbar left={...} right={...} />` 放置月份切换（上月/下月/当前月份）
- content: 使用 `ContentCard` 承载月历网格（保留现有网格结构），选中日期详情维持第二张 Card

- [ ] **Step 5: 处理 ledger 缺失/未加载状态**

当 `!ledger_id` 或 `ledgerFromUrl` 为空：
- 用 `PageLayout` + `Empty description="账本不存在或未加载"` 显示

---

### Task 2: AppLayout 标题映射补齐“日历视图”，并在日历页全局账本切换时导航

**Files:**
- Modify: `frontend/src/pages/AppLayout.tsx`

- [ ] **Step 1: 修正 pageTitle 计算：日历路由优先匹配**

规则：
- 若 `location.pathname` 满足 `^/ledgers/[^/]+/calendar$`，标题强制为“日历视图”
- 否则沿用 `routeTitleMap + startsWith` 逻辑

- [ ] **Step 2: 修改 Header Select 的 onChange**

规则：
- 仍然 `setCurrentLedger(ledger)`
- 如果当前在日历视图路由：`navigate(/ledgers/${id}/calendar)`
- 其它路由：仅 set store，不导航

---

### Task 3: 增加 CalendarViewPage 最小渲染测试（保障 PageLayout 接入与“移除页内 ledger select”）

**Files:**
- Test: `frontend/src/__tests__/pages/CalendarViewPage.test.tsx`

- [ ] **Step 1: 写一个最小渲染用例**

覆盖点：
- 路由为 `/ledgers/ledger-1/calendar` 时能渲染“日历视图”
- 页面内不存在 Antd Select（用于账本选择）

测试骨架（按现有测试风格）：

```tsx
import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useAppStore } from '../../store/appStore';
import CalendarViewPage from '../../pages/CalendarViewPage';

describe('CalendarViewPage', () => {
  it('renders with PageLayout and no in-page ledger select', () => {
    useAppStore.setState({
      ledgers: [{ id: 'ledger-1', name: '测试账本', icon: '📒', base_currency: 'CNY' } as any],
      currentLedger: { id: 'ledger-1', name: '测试账本', icon: '📒', base_currency: 'CNY' } as any,
    } as any);

    const { container } = render(
      <ConfigProvider locale={zhCN}>
        <AntApp>
          <MemoryRouter initialEntries={['/ledgers/ledger-1/calendar']}>
            <Routes>
              <Route path="/ledgers/:ledger_id/calendar" element={<CalendarViewPage />} />
            </Routes>
          </MemoryRouter>
        </AntApp>
      </ConfigProvider>,
    );

    expect(screen.getByText('日历视图')).toBeInTheDocument();
    expect(container.querySelector('.ant-select')).toBeNull();
  });
});
```

- [ ] **Step 2: 跑测试并确认通过**

Run:

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

---

### Task 4: 完整验证

- [ ] **Step 1: Run tests**

```bash
cd /workspace/frontend
npm test
```

Expected: PASS

- [ ] **Step 2: Run build**

```bash
cd /workspace/frontend
npm run build
```

Expected: build success

---

## Plan Self-Review

- Spec coverage：PageLayout/PageTitle/PageToolbar 接入、移除页内 ledger select、URL ledger_id 权威、AppLayout 标题与 ledger 切换导航均有对应任务。
- Placeholder scan：无 TODO/TBD；每步含明确文件与可执行命令。
- Type consistency：沿用现有 store/useAppStore 与 react-router-dom 路由结构，不引入新依赖。
