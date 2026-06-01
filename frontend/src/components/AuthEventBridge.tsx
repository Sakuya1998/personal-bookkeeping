import React, { useEffect } from 'react';
import { App as AntApp } from 'antd';
import { useLocation, useNavigate } from 'react-router-dom';

type UnauthorizedDetail = { next?: string };

const AuthEventBridge: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { message } = AntApp.useApp();

  useEffect(() => {
    const onUnauthorized = (e: Event) => {
      const detail = (e as CustomEvent<UnauthorizedDetail>).detail;
      const next = detail?.next || `${location.pathname}${location.search}`;
      const safeNext = next.startsWith('/') ? next : '/';
      message.error('登录已过期，请重新登录');
      navigate(`/login?next=${encodeURIComponent(safeNext)}`, { replace: true });
    };

    window.addEventListener('auth:unauthorized', onUnauthorized as EventListener);
    return () => window.removeEventListener('auth:unauthorized', onUnauthorized as EventListener);
  }, [navigate, message, location.pathname, location.search]);

  return null;
};

export default AuthEventBridge;

