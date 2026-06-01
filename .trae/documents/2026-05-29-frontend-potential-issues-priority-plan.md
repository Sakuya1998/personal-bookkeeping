# 前端潜在问题（按优先级）处理计划

> **目标：**按优先级修复当前前端中已识别的潜在问题，覆盖构建配置正确性、PWA 缓存安全性、401 鉴权失效的跳转体验与可回跳登录流程，并用现有 Vitest 测试兜底。

## Summary

- 修复 Vite 构建分包配置项键名错误，确保 `manualChunks` 生效并减少生产包体积风险。
- 禁用 PWA 对 `/api` 的运行时缓存，避免鉴权响应/错误被缓存造成跨用户污染与登出后读旧数据。
- 将 401 处理从“硬跳转”改为“路由内跳转 + 回跳”，并同步改造登录页支持 `next` 参数，提升体验并降低状态丢失。

## Current State Analysis（基于仓库现状）

### P0：构建正确性

- [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts) 中 `build.rolldownOptions` 疑似为拼写错误；Vite/Rollup 对应字段应为 `rollupOptions`。当前写法可能导致 `manualChunks` 完全不生效，进而增加首屏包体积与性能风险。

### P0：PWA 缓存安全/一致性

- [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts) 的 Workbox `runtimeCaching` 将所有匹配 `/api/` 的请求按 `NetworkFirst` 缓存（见 `cacheName: 'api-cache'`）。这会把带鉴权的响应也纳入缓存，存在以下风险：
  - 登出/切换用户后仍读取到旧缓存数据；
  - 缓存住 401/500 等错误响应导致排障困难；
  - 缓存 key 未区分 Authorization，存在跨会话污染可能。

### P1：鉴权失效与跳转体验

- [frontend/src/api/client.ts](file:///workspace/frontend/src/api/client.ts) 在 401 时直接 `window.location.href='/login'`，会触发整页刷新并丢失 SPA 状态；且无法“登录后回到原页面”。
- [frontend/src/pages/LoginPage.tsx](file:///workspace/frontend/src/pages/LoginPage.tsx) 登录/注册成功后固定 `navigate('/')`，不支持回跳。
- token 来源分散：store 初始化从 localStorage 取 token（[appStore.ts](file:///workspace/frontend/src/store/appStore.ts)），axios 请求拦截器也直接读 localStorage（[client.ts](file:///workspace/frontend/src/api/client.ts)），后续扩展（例如更细粒度登出/多 Tab 同步）容易出现不一致。

## Assumptions & Decisions

- 仅改造前端，不引入新依赖、不引入 refresh token 机制（需要后端配合）。
- PWA 对 `/api` 的运行时缓存：选择“禁用 /api 缓存”（用户已确认）。
- 401 行为：选择“路由内跳转 + 保留回跳 next”（用户已确认）。
- `next` 仅允许站内路径（必须以 `/` 开头），避免开放重定向。

## Proposed Changes（按优先级）

### P0：修复 Vite 分包配置键名错误

**Files:**
- Modify: [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts)

**Change:**
- 将 `build.rolldownOptions` 更正为 `build.rollupOptions`，确保 `manualChunks` 生效。

**Target code (replace build block):**

```ts
build: {
  chunkSizeWarningLimit: 400,
  rollupOptions: {
    output: {
      manualChunks: (id: string) => {
        if (id.includes('node_modules/echarts')) return 'echarts'
        if (id.includes('node_modules/antd') || id.includes('node_modules/@ant-design')) return 'antd'
        if (id.includes('node_modules/react')) return 'vendor'
      },
    },
  },
},
```

### P0：禁用 PWA 对 /api 的 runtimeCaching

**Files:**
- Modify: [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts)

**Change:**
- 移除 `workbox.runtimeCaching` 中对 `/api` 的缓存规则（保留 precache 的 `globPatterns`）。

**Target code (replace workbox block):**

```ts
workbox: {
  globPatterns: ['**/*.{js,css,html,svg,png,ico,json}'],
},
```

### P1：统一 401 处理为“事件通知 + 路由内跳转（含 next）”

**Files:**
- Modify: [frontend/src/api/client.ts](file:///workspace/frontend/src/api/client.ts)
- Create: `/workspace/frontend/src/components/AuthEventBridge.tsx`
- Modify: [frontend/src/App.tsx](file:///workspace/frontend/src/App.tsx)

**Design:**
- axios 拦截器不直接导航，而是：
  - 清理 token（调用 store.logout），并触发一次 `CustomEvent('auth:unauthorized')`；
  - 事件 detail 包含 `next`（默认取 `window.location.pathname + window.location.search`）。
- UI 层在 Router 内挂载 `AuthEventBridge` 监听该事件，收到后：
  - 使用 `navigate('/login?next=...')` 做 SPA 内跳转；
  - 可选提示（使用 antd 的 message）。
- 为避免多请求并发触发多次跳转，在 client 模块内用布尔锁防抖（仅第一次 401 触发事件）。

**Target code: `src/api/client.ts`（整体替换为以下结构）**

```ts
import axios from 'axios'
import { useAppStore } from '../store/appStore'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const token = useAppStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let handling401 = false

client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401 && !handling401) {
      handling401 = true
      useAppStore.getState().logout()
      window.dispatchEvent(
        new CustomEvent('auth:unauthorized', {
          detail: { next: window.location.pathname + window.location.search },
        }),
      )
    }
    return Promise.reject(err)
  },
)

export function resetUnauthorizedHandlingForTests() {
  handling401 = false
}

export default client
```

**Target code: `src/components/AuthEventBridge.tsx`（新增文件）**

```tsx
import React, { useEffect } from 'react'
import { App as AntApp } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'

type UnauthorizedDetail = { next?: string }

const AuthEventBridge: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { message } = AntApp.useApp()

  useEffect(() => {
    const onUnauthorized = (e: Event) => {
      const detail = (e as CustomEvent<UnauthorizedDetail>).detail
      const next = detail?.next || `${location.pathname}${location.search}`
      const safeNext = next.startsWith('/') ? next : '/'
      message.error('登录已过期，请重新登录')
      navigate(`/login?next=${encodeURIComponent(safeNext)}`, { replace: true })
    }

    window.addEventListener('auth:unauthorized', onUnauthorized as EventListener)
    return () => window.removeEventListener('auth:unauthorized', onUnauthorized as EventListener)
  }, [navigate, message, location.pathname, location.search])

  return null
}

export default AuthEventBridge
```

**Target code: `src/App.tsx`（在 Router 内注入桥接组件）**

- 增加 `import AuthEventBridge from './components/AuthEventBridge'`
- 在 `<BrowserRouter>` 内、`<Routes>` 之前渲染 `<AuthEventBridge />`

```tsx
<BrowserRouter>
  <AuthEventBridge />
  <ErrorBoundary>
    <Suspense fallback={<PageLoading />}>
      ...
    </Suspense>
  </ErrorBoundary>
</BrowserRouter>
```

### P1：登录/注册成功后支持 next 回跳

**Files:**
- Modify: [frontend/src/pages/LoginPage.tsx](file:///workspace/frontend/src/pages/LoginPage.tsx)

**Change:**
- 使用 `useSearchParams()` 读取 `next`，登录/注册成功后 `navigate(safeNext, { replace: true })`。
- `safeNext` 仅允许以 `/` 开头的站内路径，否则回落到 `/`。

**Target code（关键改动点）**

```tsx
import { useNavigate, useSearchParams } from 'react-router-dom'

const [searchParams] = useSearchParams()
const next = searchParams.get('next') || '/'
const safeNext = next.startsWith('/') ? next : '/'

...
navigate(safeNext, { replace: true })
```

## Testing & Verification

### 单元测试更新（Vitest）

**Files:**
- Modify: [frontend/src/__tests__/api/client.test.ts](file:///workspace/frontend/src/__tests__/api/client.test.ts)
- Create: `/workspace/frontend/src/__tests__/components/AuthEventBridge.test.tsx`
- Create: `/workspace/frontend/src/__tests__/pages/LoginPage.test.tsx`

**client.test.ts（需要更新断言）**
- request interceptor：从 store 取 token（保持“先写 localStorage 再 import client”仍可通过，因为 store 初始化读取 localStorage）。
- response error handler：
  - 401 时：应清除 localStorage token，触发 `auth:unauthorized` 事件；不再断言 `window.location.href`。
  - 增加对 `resetUnauthorizedHandlingForTests()` 的调用，避免测试互相影响。

**AuthEventBridge.test.tsx（核心行为）**
- 用 `MemoryRouter` 包住 `AuthEventBridge`，初始路径设置为 `/transactions?x=1`。
- `window.dispatchEvent(new CustomEvent('auth:unauthorized', { detail: { next: '/transactions?x=1' } }))` 后：
  - 断言路由跳转到 `/login?next=%2Ftransactions%3Fx%3D1`（以 UI 变化或 router location 断言为准）。

**LoginPage.test.tsx（回跳行为）**
- 使用 `MemoryRouter` + 初始 URL `/login?next=/transactions` 渲染 LoginPage。
- mock `client.post('/auth/login')` 返回 token/user，触发表单提交后：
  - 断言 `navigate('/transactions', { replace: true })` 生效（可通过渲染一个 `<Routes>` + `<Route path="/transactions" element={<div>OK</div>} />` 断言 OK 出现）。
- 增加一条安全性用例：`next=https://evil.com` 或 `next=//evil.com` 时回落 `/`。

### 命令级验证（本地/CI）

在 `frontend/` 目录下执行：

- `npm test`
- `npm run lint`
- `npm run build`

验收标准：
- `npm run build` 无警告/错误，且产物分包（`antd`/`echarts`/`vendor`）在 `dist/assets` 里能观察到独立 chunk（以文件名包含对应 chunk 名或构建日志为准）。
- Service Worker 不再对 `/api` 请求建立 `api-cache` 运行时缓存（可在浏览器 DevTools Application → Cache Storage 侧验证）。
- 401 发生时不整页刷新，跳转到 `/login?next=...`，登录成功后回到原页面。

---

## Task Breakdown（可直接执行的步骤清单）

### Task 1（P0）：修复 Vite rollupOptions 键名

**Files:**
- Modify: [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts)

- [ ] 将 `rolldownOptions` 改为 `rollupOptions`，保持现有 `manualChunks` 逻辑不变

### Task 2（P0）：移除 /api runtimeCaching

**Files:**
- Modify: [frontend/vite.config.ts](file:///workspace/frontend/vite.config.ts)

- [ ] 删除 `runtimeCaching` 数组中针对 `/api/` 的条目（或整体移除 runtimeCaching），保留 precache `globPatterns`

### Task 3（P1）：401 事件化 + 路由内跳转

**Files:**
- Modify: [frontend/src/api/client.ts](file:///workspace/frontend/src/api/client.ts)
- Create: `/workspace/frontend/src/components/AuthEventBridge.tsx`
- Modify: [frontend/src/App.tsx](file:///workspace/frontend/src/App.tsx)

- [ ] client：请求拦截器从 `useAppStore.getState().token` 注入 Authorization
- [ ] client：401 时调用 `logout()` 并派发 `auth:unauthorized`（带 next），加 `handling401` 锁
- [ ] App：挂载 `AuthEventBridge` 监听事件并 `navigate('/login?next=...')`

### Task 4（P1）：LoginPage 支持 next 回跳

**Files:**
- Modify: [frontend/src/pages/LoginPage.tsx](file:///workspace/frontend/src/pages/LoginPage.tsx)

- [ ] 读取 `next`，校验为站内路径后在登录/注册成功时 replace 跳转

### Task 5：测试与验证

**Files:**
- Modify: [frontend/src/__tests__/api/client.test.ts](file:///workspace/frontend/src/__tests__/api/client.test.ts)
- Create: `/workspace/frontend/src/__tests__/components/AuthEventBridge.test.tsx`
- Create: `/workspace/frontend/src/__tests__/pages/LoginPage.test.tsx`

- [ ] 更新 client 单测以覆盖“401 触发事件而非硬跳转”
- [ ] 新增 AuthEventBridge 单测覆盖事件驱动跳转与 next 编码
- [ ] 新增 LoginPage 单测覆盖 next 回跳与安全回落
- [ ] 运行 `frontend` 的 `npm test && npm run lint && npm run build`

