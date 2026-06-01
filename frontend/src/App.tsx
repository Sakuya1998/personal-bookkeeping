import React, { Suspense } from 'react';
import { Spin } from 'antd';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { useAppStore } from './store/appStore';
import ErrorBoundary from './components/ErrorBoundary';
import AuthEventBridge from './components/AuthEventBridge';
import AppLayout from './pages/AppLayout';
const LoginPage = React.lazy(() => import('./pages/LoginPage'));
const TransactionsPage = React.lazy(() => import('./pages/TransactionsPage'));
const LedgersPage = React.lazy(() => import('./pages/LedgersPage'));
const CategoriesPage = React.lazy(() => import('./pages/CategoriesPage'));
const ExchangeRatesPage = React.lazy(() => import('./pages/ExchangeRatesPage'));
const SettingsPage = React.lazy(() => import('./pages/SettingsPage'));
const RecurringPage = React.lazy(() => import('./pages/RecurringPage'));
const BudgetPage = React.lazy(() => import('./pages/BudgetPage'));

const DashboardPage = React.lazy(() => import('./pages/DashboardPage'));
const CalendarViewPage = React.lazy(() => import('./pages/CalendarViewPage'));

const PageLoading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
    <Spin size="large" />
  </div>
);

dayjs.locale('zh-cn');

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = useAppStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <AntApp>
        <BrowserRouter>
          <AuthEventBridge />
          <ErrorBoundary>
            <Suspense fallback={<PageLoading />}>
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/" element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
                  <Route index element={<DashboardPage />} />
                  <Route path="transactions" element={<TransactionsPage />} />
                  <Route path="ledgers" element={<LedgersPage />} />
                  <Route path="categories" element={<CategoriesPage />} />
                  <Route path="exchange-rates" element={<ExchangeRatesPage />} />
                  <Route path="settings" element={<SettingsPage />} />
                  <Route path="recurring" element={<RecurringPage />} />
                  <Route path="budgets" element={<BudgetPage />} />
                  <Route path="ledgers/:ledger_id/calendar" element={<CalendarViewPage />} />
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </Suspense>
          </ErrorBoundary>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
