import React from 'react';
import { Card, Form, Input, Button, message } from 'antd';
import client from '../api/client';
import { ApiResponse, User } from '../api/types';
import { useAppStore } from '../store/appStore';

const SettingsPage: React.FC = () => {
  const { user, setUser } = useAppStore();

  const handleSubmit = async (values: any) => {
    // Settings page - future implementation for user profile updates
    message.info('设置保存功能开发中');
  };

  return (
    <Card title="个人设置">
      <Form layout="vertical" style={{ maxWidth: 400 }} initialValues={{ username: user?.username, email: user?.email }} onFinish={handleSubmit}>
        <Form.Item name="username" label="用户名">
          <Input disabled />
        </Form.Item>
        <Form.Item name="email" label="邮箱" rules={[{ type: 'email' }]}>
          <Input />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit">保存</Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default SettingsPage;
