import React, { useEffect, useState } from 'react';
import { Layout, Menu, Select, Button, Dropdown, message } from 'antd';
import {
  DashboardOutlined, WalletOutlined, TransactionOutlined, AppstoreOutlined,
  DollarOutlined, SettingOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import client from '../api/client';
import { ApiResponse, Ledger, User } from '../api/types';
import { useAppStore } from '../store/appStore';

const { Header, Sider, Content } = Layout;

const AppLayout: React.FC = () => {
  const { user, setUser, setLedgers, ledgers, currentLedger, setCurrentLedger, logout, token } = useAppStore();
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

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
  }, []);

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/transactions', icon: <TransactionOutlined />, label: '交易记录' },
    { key: '/ledgers', icon: <WalletOutlined />, label: '账本管理' },
    { key: '/categories', icon: <AppstoreOutlined />, label: '分类管理' },
    { key: '/exchange-rates', icon: <DollarOutlined />, label: '汇率管理' },
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
        <div style={{ height: 32, margin: 16, color: '#fff', fontWeight: 'bold', fontSize: collapsed ? 14 : 18, textAlign: 'center' }}>
          {collapsed ? '📒' : '📒 记账'}
        </div>
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
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {ledgers.length > 0 && (
              <Select
                value={currentLedger?.id}
                onChange={(id) => setCurrentLedger(ledgers.find(l => l.id === id) || null)}
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
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
