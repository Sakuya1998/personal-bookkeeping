import React, { Suspense } from 'react';
import { Spin } from 'antd';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { useAppStore } from './store/appStore';
import AppLayout from './pages/AppLayout';
import LoginPage from './pages/LoginPage';
import TransactionsPage from './pages/TransactionsPage';
import LedgersPage from './pages/LedgersPage';
import CategoriesPage from './pages/CategoriesPage';
import ExchangeRatesPage from './pages/ExchangeRatesPage';
import SettingsPage from './pages/SettingsPage';

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
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
              <Route index element={<Suspense fallback={<PageLoading />}><DashboardPage /></Suspense>} />
              <Route path="transactions" element={<TransactionsPage />} />
              <Route path="ledgers" element={<LedgersPage />} />
              <Route path="categories" element={<CategoriesPage />} />
              <Route path="exchange-rates" element={<ExchangeRatesPage />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="ledgers/:ledger_id/calendar" element={<Suspense fallback={<PageLoading />}><CalendarViewPage /></Suspense>} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
