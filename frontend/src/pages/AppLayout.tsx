import React, { useEffect, useState } from 'react';
import { Layout, Menu, Select, Button, Dropdown } from 'antd';
import {
  DashboardOutlined, WalletOutlined, TransactionOutlined, AppstoreOutlined,
  DollarOutlined, SettingOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined,
  SyncOutlined, FundOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import client from '../api/client';
import { ApiResponse, Ledger, User } from '../api/types';
import { useAppStore } from '../store/appStore';
import Brand from '../components/layout/Brand';

const { Header, Sider, Content } = Layout;

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

const AppLayout: React.FC = () => {
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

  const pageTitle = isCalendarView
    ? '日历视图'
    : routeTitleMap[location.pathname]
      || Object.entries(routeTitleMap).find(([path]) => path !== '/' && location.pathname.startsWith(path))?.[1]
      || '个人记账';

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
    { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/transactions', icon: <TransactionOutlined />, label: '交易记录' },
    { key: '/ledgers', icon: <WalletOutlined />, label: '账本管理' },
    { key: '/categories', icon: <AppstoreOutlined />, label: '分类管理' },
    { key: '/exchange-rates', icon: <DollarOutlined />, label: '汇率管理' },
    { key: '/recurring', icon: <SyncOutlined />, label: '周期规则' },
    { key: '/budgets', icon: <FundOutlined />, label: '预算管理' },
    { key: '/settings', icon: <SettingOutlined />, label: '设置' },
  ];

  const userMenu = {
    items: [
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
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
              <span style={{ cursor: 'pointer' }}>{user?.username || '用户'}</span>
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
