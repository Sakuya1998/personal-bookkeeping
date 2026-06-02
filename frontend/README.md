# 个人记账 — 前端

> 文档版本: v4.0

基于 **React 19** + **TypeScript** + **Vite** 构建的个人记账应用前端，使用 **Ant Design 6** 提供 UI 组件。

---

## 技术栈

| 类别 | 选型 |
|------|------|
| 框架 | React 19 (Strict Mode) |
| 语言 | TypeScript 5.x |
| 构建 | Vite 6.x |
| UI 组件 | Ant Design 6.x |
| 状态管理 | Zustand |
| 路由 | React Router 7.x |
| 图表 | ECharts 5.x |
| 国际化 | react-i18next |
| HTTP 客户端 | Axios |
| 测试框架 | Vitest + @testing-library/react |
| API Mock | MSW |
| PWA | vite-plugin-pwa (Workbox) |

---

## 可用脚本

```bash
# 启动开发服务器 (默认端口 5173)
npm run dev

# 生产构建
npm run build

# 本地预览生产构建
npm run preview

# 运行测试
npm test

# 测试覆盖率
npm run test:coverage

# 类型检查
npx tsc --noEmit

# Lint
npm run lint
```

---

## 项目结构

```
frontend/
├── public/
│   ├── icons/                  # PWA 图标
│   ├── locales/                # i18n 翻译文件
│   │   ├── zh-CN/
│   │   │   └── translation.json    # 中文翻译 (~500 key)
│   │   └── en-US/
│   │       └── translation.json    # 英文翻译
│   ├── manifest.webmanifest    # PWA 清单
│   └── sw.js                   # Service Worker (自动生成)
├── src/
│   ├── __tests__/              # 前端测试
│   │   ├── api/
│   │   │   └── client.test.ts       # Axios 客户端 (153 行)
│   │   ├── store/
│   │   │   └── appStore.test.ts     # Zustand 状态管理 (244 行)
│   │   └── utils/
│   │       └── currency.test.ts     # 货币工具函数 (54 行)
│   ├── api/                    # API 客户端
│   │   └── client.ts                # Axios 实例 + 拦截器
│   ├── components/             # 通用组件
│   │   ├── i18n/               # 国际化组件
│   │   │   └── LanguageSelector.tsx  # 语言切换
│   │   ├── ledger/             # 账本相关组件
│   │   └── layout/             # 布局组件
│   ├── hooks/                  # 自定义 Hook
│   │   ├── useRole.ts               # 角色权限判断
│   │   └── useTranslation.ts        # i18n Hook
│   ├── i18n/                   # i18n 配置
│   │   └── index.ts                 # react-i18next 初始化
│   ├── pages/                  # 页面组件
│   │   ├── DashboardPage/           # 仪表盘
│   │   ├── TransactionsPage/        # 交易管理
│   │   ├── LedgerPage/              # 账本管理
│   │   ├── MemberPage/              # 成员管理 ✨ v4.0
│   │   ├── CalendarViewPage/        # 日历视图
│   │   ├── SettingsPage/            # 设置
│   │   ├── BudgetPage/              # 预算管理
│   │   ├── RecurringPage/           # 周期性交易
│   │   ├── ReportPage/              # 报表
│   │   ├── AnnualReportPage/        # 年度财务报告 ✨ v4.0
│   │   └── RecycledBinPage/         # 回收站 ✨ v4.0
│   ├── store/                  # Zustand 状态
│   │   └── appStore.ts              # 全局状态
│   ├── utils/                  # 工具函数
│   │   └── currency.ts             # 货币格式化
│   ├── App.tsx                 # 根组件
│   ├── main.tsx                # 入口
│   └── routes.tsx              # 路由配置
├── index.html
├── vite.config.ts              # Vite 配置
├── vitest.config.ts            # 测试配置
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
└── package.json
```

---

## 国际化 (i18n)

使用 **react-i18next** 实现完整中英双语支持。

- 翻译文件位于 `public/locales/{zh-CN,en-US}/translation.json`
- 语言选择器 (`LanguageSelector`) 位于导航栏
- 用户偏好持久化到 `localStorage`
- 默认跟随浏览器语言 (fallback: zh-CN)

### 使用方式

```tsx
import { useTranslation } from 'react-i18next';

function MyComponent() {
  const { t } = useTranslation();
  return <h1>{t('dashboard.title')}</h1>;
}
```

---

## 测试

| 配置 | 说明 |
|------|------|
| 框架 | Vitest |
| React 测试 | @testing-library/react |
| API Mock | MSW |
| 覆盖率 | `npm run test:coverage` |

### 当前测试覆盖

- `store/appStore.test.ts` — Zustand 状态管理 (244 行)
- `api/client.test.ts` — Axios 请求/响应拦截器 (153 行)
- `utils/currency.test.ts` — 货币格式化/金额计算 (54 行)

---

## 构建与部署

```bash
# 构建
npm run build

# 产物输出到 dist/
# 静态文件部署到 nginx / CDN
# API 代理在 nginx 层配置 /api -> 后端服务
```

Vite 配置了：
- 代码分割 (路由级懒加载)
- PWA (Service Worker + 桌面图标)
- ESLint + TypeScript 检查
- CSS 压缩
- 资源指纹
