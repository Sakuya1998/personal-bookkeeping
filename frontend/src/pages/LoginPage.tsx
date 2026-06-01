import React, { useEffect, useRef, useState } from 'react';
import { Card, Form, Input, Button, Tabs, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import client from '../api/client';
import { ApiResponse, AuthResponse } from '../api/types';
import { useAppStore } from '../store/appStore';
import { useNavigate, useSearchParams } from 'react-router-dom';

const LoginPage: React.FC = () => {
  const [tab, setTab] = useState<'login' | 'register'>('login');
  const [loading, setLoading] = useState(false);
  const mountedRef = useRef(true);
  const { setToken, setUser } = useAppStore();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const next = searchParams.get('next') || '/';
  const safeNext = next.startsWith('/') ? next : '/';

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const onLogin = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const res = await client.post<ApiResponse<AuthResponse>>('/auth/login', values);
      setToken(res.data.data.token);
      setUser(res.data.data.user);
      message.success('登录成功');
      navigate(safeNext, { replace: true });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '登录失败');
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  };

  const onRegister = async (values: { username: string; email: string; password: string }) => {
    setLoading(true);
    try {
      const res = await client.post<ApiResponse<AuthResponse>>('/auth/register', values);
      setToken(res.data.data.token);
      setUser(res.data.data.user);
      message.success('注册成功');
      navigate(safeNext, { replace: true });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '注册失败');
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  };

  const items = [
    {
      key: 'login',
      label: '登录',
      children: (
        <Form onFinish={onLogin} layout="vertical">
          <Form.Item label="用户名" name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="例如：alice" size="large" autoComplete="username" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" size="large" autoComplete="current-password" />
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

  return (
    <div style={{ minHeight: '100dvh', display: 'grid', placeItems: 'center', padding: 16 }}>
      <Card style={{ width: 420, borderRadius: 10 }}>
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 20, fontWeight: 600 }}>个人记账</div>
          <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>快速记录每一笔收支</div>
        </div>
        <Tabs activeKey={tab} onChange={(k) => setTab(k as 'login' | 'register')} items={items} centered />
      </Card>
    </div>
  );
};

export default LoginPage;
