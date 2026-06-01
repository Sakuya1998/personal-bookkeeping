import React, { useEffect, useState } from 'react';
import { Layout, Menu, Select, Button, Dropdown } from 'antd';
import {
  DashboardOutlined, WalletOutlined, TransactionOutlined, AppstoreOutlined,
  DollarOutlined, SettingOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined,
  SyncOutlined, FundOutlined, TagsOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import client from '../api/client';
import { ApiResponse, Ledger, User } from '../api/types';
import { useAppStore } from '../store/appStore';
import Brand from '../components/layout/Brand';
import { useTranslation } from 'react-i18next';

const { Header, Sider, Content } = Layout;

const AppLayout: React.FC = () => {
  const { t } = useTranslation();
  const user = useAppStore(s => s.user);
  const token = useAppStore(s => s.token);
  const ledgers = useAppStore(s => s.ledgers);
  const currentLedger = useAppStore(s => s.currentLedger);
  const setUser = useAppStore(s => s.setUser);
  const setLedgers = useAppStore(s => s.setLedgers);
  const setCurrentLedger = useAppStore(s => s.setCurrentLedger);
  const logout = useAppStore(s => s.logout);
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const isCalendarView = /^\/ledgers\/[^/]+\/calendar$/.test(location.pathname);

  const routeTitleMap: Record<string, string> = {
    '/': t('nav.dashboard'),
    '/transactions': t('nav.transactions'),
    '/ledgers': t('nav.ledgers'),
    '/categories': t('nav.categories'),
    '/exchange-rates': t('nav.exchangeRates'),
    '/recurring': t('nav.recurring'),
    '/tag-stats': t('nav.tagStats'),
    '/budgets': t('nav.budgets'),
    '/settings': t('nav.settings'),
  };

  const pageTitle = isCalendarView
    ? t('nav.calendar')
    : routeTitleMap[location.pathname]
      || Object.entries(routeTitleMap).find(([path]) => path !== '/' && location.pathname.startsWith(path))?.[1]
      || t('app.title');

  useEffect(() => {
    if (!token) { navigate('/login'); return; }
    // Load user info
    client.get<ApiResponse<User>>('/auth/me').then((res) => {
      setUser(res.data.data);
    }).catch(() => {
      logout();
      navigate('/login');
    });
    // Load ledgers
    client.get<ApiResponse<Ledger[]>>('/ledgers').then((res) => {
      setLedgers(res.data.data);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: t('nav.dashboard') },
    { key: '/transactions', icon: <TransactionOutlined />, label: t('nav.transactions') },
    { key: '/ledgers', icon: <WalletOutlined />, label: t('nav.ledgers') },
    { key: '/categories', icon: <AppstoreOutlined />, label: t('nav.categories') },
    { key: '/exchange-rates', icon: <DollarOutlined />, label: t('nav.exchangeRates') },
    { key: '/recurring', icon: <SyncOutlined />, label: t('nav.recurring') },
    { key: '/budgets', icon: <FundOutlined />, label: t('nav.budgets') },
    { key: '/tag-stats', icon: <TagsOutlined />, label: t('nav.tagStats') },
    { key: '/settings', icon: <SettingOutlined />, label: t('nav.settings') },
  ];

  const userMenu = {
    items: [
      { key: 'logout', icon: <LogoutOutlined />, label: t('nav.logout'), danger: true },
    ],
    onClick: ({ key }: { key: string }) => {
      if (key === 'logout') {
        logout();
        navigate('/login');
      }
    },
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} breakpoint="lg" onBreakpoint={(broken) => setCollapsed(broken)}>
        <Brand collapsed={collapsed} />
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid #f0f0f0' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
            />
            <div style={{ fontSize: 16, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{pageTitle}</div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {ledgers.length > 0 && (
              <Select
                value={currentLedger?.id}
                onChange={(id) => {
                  const ledger = ledgers.find((l) => l.id === id) || null;
                  setCurrentLedger(ledger);
                  if (isCalendarView) {
                    navigate(`/ledgers/${id}/calendar`);
                  }
                }}
                style={{ width: 180 }}
                options={ledgers.map(l => ({ label: `${l.icon || ''} ${l.name}`, value: l.id }))}
                size="small"
              />
            )}
            <Dropdown menu={userMenu}>
              <span style={{ cursor: 'pointer' }}>{user?.username || t('nav.user')}</span>
            </Dropdown>
          </div>
        </Header>
        <Content style={{ padding: 0 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
