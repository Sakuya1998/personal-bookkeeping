import React, { useEffect, useRef, useState } from 'react';
import { Card, Form, Input, Button, Tabs, message } from 'antd';
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import client from '../api/client';
import { ApiResponse, AuthResponse } from '../api/types';
import { useAppStore } from '../store/appStore';
import { useNavigate, useSearchParams } from 'react-router-dom';

const LoginPage: React.FC = () => {
  const { t } = useTranslation();
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
      message.success(t('auth.loginSuccess'));
      navigate(safeNext, { replace: true });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('auth.loginFailed'));
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
      message.success(t('auth.registerSuccess'));
      navigate(safeNext, { replace: true });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('auth.registerFailed'));
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  };

  const items = [
    {
      key: 'login',
      label: t('auth.login'),
      children: (
        <Form onFinish={onLogin} layout="vertical">
          <Form.Item label={t('auth.username')} name="username" rules={[{ required: true, message: t('auth.usernameRequired') }]}>
            <Input prefix={<UserOutlined />} placeholder={t('auth.usernamePlaceholder')} size="large" autoComplete="username" />
          </Form.Item>
          <Form.Item label={t('auth.password')} name="password" rules={[{ required: true, message: t('auth.passwordRequired') }]}>
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.passwordPlaceholder')} size="large" autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block size="large">
            {t('auth.login')}
          </Button>
        </Form>
      ),
    },
    {
      key: 'register',
      label: t('auth.register'),
      children: (
        <Form onFinish={onRegister} layout="vertical">
          <Form.Item label={t('auth.username')} name="username" rules={[{ required: true, min: 2, message: t('auth.usernameMinLength') }]}>
            <Input prefix={<UserOutlined />} placeholder={t('auth.usernamePlaceholder')} size="large" autoComplete="username" />
          </Form.Item>
          <Form.Item label={t('auth.email')} name="email" rules={[{ required: true, type: 'email', message: t('auth.emailInvalid') }]}>
            <Input prefix={<MailOutlined />} placeholder={t('auth.emailPlaceholder')} size="large" autoComplete="email" />
          </Form.Item>
          <Form.Item label={t('auth.password')} name="password" rules={[{ required: true, min: 6, message: t('auth.passwordMinLength') }]}>
            <Input.Password prefix={<LockOutlined />} placeholder={t('auth.passwordPlaceholder')} size="large" autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block size="large">
            {t('auth.register')}
          </Button>
        </Form>
      ),
    },
  ];

  return (
    <div style={{ minHeight: '100dvh', display: 'grid', placeItems: 'center', padding: 16 }}>
      <Card style={{ width: 420, borderRadius: 10 }}>
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 20, fontWeight: 600 }}>{t('auth.appTitle')}</div>
          <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>{t('auth.appSubtitle')}</div>
        </div>
        <Tabs activeKey={tab} onChange={(k) => setTab(k as 'login' | 'register')} items={items} centered />
      </Card>
    </div>
  );
};

export default LoginPage;
