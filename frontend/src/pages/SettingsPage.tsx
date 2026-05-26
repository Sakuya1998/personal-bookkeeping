import React, { useState } from 'react';
import { Card, Form, Input, Button, message, Divider, Row, Col } from 'antd';
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons';
import { useAppStore } from '../store/appStore';
import client from '../api/client';
import { ApiResponse, User } from '../api/types';

const SettingsPage: React.FC = () => {
  const { user, setUser } = useAppStore();
  const [pwLoading, setPwLoading] = useState(false);
  const [emailLoading, setEmailLoading] = useState(false);
  const [pwForm] = Form.useForm();
  const [emailForm] = Form.useForm();

  const handleChangePassword = async (values: { old_password: string; new_password: string; confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) {
      message.error('两次输入的新密码不一致');
      return;
    }
    setPwLoading(true);
    try {
      await client.put('/auth/password', {
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success('密码修改成功');
      pwForm.resetFields();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '密码修改失败');
    } finally {
      setPwLoading(false);
    }
  };

  const handleChangeEmail = async (values: { email: string }) => {
    setEmailLoading(true);
    try {
      const res = await client.put<ApiResponse<User>>('/auth/email', { email: values.email });
      message.success('邮箱修改成功');
      setUser(res.data.data);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '邮箱修改失败');
    } finally {
      setEmailLoading(false);
    }
  };

  return (
    <Row gutter={[24, 24]}>
      <Col xs={24} md={12}>
        <Card title="个人信息">
          <Form layout="vertical">
            <Form.Item label="用户名">
              <Input prefix={<UserOutlined />} value={user?.username || ''} disabled />
            </Form.Item>
            <Form.Item label="当前邮箱">
              <Input prefix={<MailOutlined />} value={user?.email || ''} disabled />
            </Form.Item>
          </Form>
        </Card>
      </Col>

      <Col xs={24} md={12}>
        <Card title="修改邮箱">
          <Form form={emailForm} layout="vertical" onFinish={handleChangeEmail} initialValues={{ email: user?.email }}>
            <Form.Item
              name="email"
              label="新邮箱"
              rules={[
                { required: true, message: '请输入新邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' },
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="new@example.com" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={emailLoading}>更新邮箱</Button>
            </Form.Item>
          </Form>

          <Divider />

          <Form form={pwForm} layout="vertical" onFinish={handleChangePassword}>
            <Form.Item
              name="old_password"
              label="当前密码"
              rules={[{ required: true, message: '请输入当前密码' }]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="输入当前密码" />
            </Form.Item>
            <Form.Item
              name="new_password"
              label="新密码"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 6, message: '密码至少 6 个字符' },
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="输入新密码" />
            </Form.Item>
            <Form.Item
              name="confirm_password"
              label="确认新密码"
              dependencies={['new_password']}
              rules={[
                { required: true, message: '请再次输入新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('new_password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="再次输入新密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={pwLoading} danger>修改密码</Button>
            </Form.Item>
          </Form>
        </Card>
      </Col>
    </Row>
  );
};

export default SettingsPage;
