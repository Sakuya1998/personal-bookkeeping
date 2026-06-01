import React, { useState } from 'react';
import { Form, Input, Button, message, Row, Col, Select } from 'antd';
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAppStore } from '../store/appStore';
import client from '../api/client';
import { ApiResponse, User } from '../api/types';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import ContentCard from '../components/layout/ContentCard';

const SettingsPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { user, setUser } = useAppStore();
  const [pwLoading, setPwLoading] = useState(false);
  const [emailLoading, setEmailLoading] = useState(false);
  const [pwForm] = Form.useForm();
  const [emailForm] = Form.useForm();

  const handleChangePassword = async (values: { old_password: string; new_password: string; confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) {
      message.error(t('auth.passwordNotMatch'));
      return;
    }
    setPwLoading(true);
    try {
      await client.put('/auth/password', {
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success(t('settings.passwordSuccess'));
      pwForm.resetFields();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('settings.passwordFailed'));
    } finally {
      setPwLoading(false);
    }
  };

  const handleChangeEmail = async (values: { email: string }) => {
    setEmailLoading(true);
    try {
      const res = await client.put<ApiResponse<User>>('/auth/email', { email: values.email });
      message.success(t('settings.emailSuccess'));
      setUser(res.data.data);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('settings.emailFailed'));
    } finally {
      setEmailLoading(false);
    }
  };

  return (
    <PageLayout header={<PageTitle title={t('settings.title')} description={t('settings.description')} />}>
      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <ContentCard title={t('settings.personalInfo')}>
            <Form layout="vertical">
              <Form.Item label={t('settings.username')}>
                <Input prefix={<UserOutlined />} value={user?.username || ''} disabled />
              </Form.Item>
              <Form.Item label={t('settings.currentEmail')}>
                <Input prefix={<MailOutlined />} value={user?.email || ''} disabled />
              </Form.Item>
            </Form>
          </ContentCard>
        </Col>

        <Col xs={24} lg={12}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <ContentCard title={t('settings.changeEmail')}>
              <Form form={emailForm} layout="vertical" onFinish={handleChangeEmail} initialValues={{ email: user?.email }}>
                <Form.Item
                  name="email"
                  label={t('settings.newEmail')}
                  rules={[
                    { required: true, message: t('settings.newEmailRequired') },
                    { type: 'email', message: t('auth.emailInvalid') },
                  ]}
                >
                  <Input prefix={<MailOutlined />} placeholder={t('settings.emailPlaceholder')} />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={emailLoading}>{t('settings.updateEmail')}</Button>
                </Form.Item>
              </Form>
            </ContentCard>

            <ContentCard title={t('settings.changePassword')}>
              <Form form={pwForm} layout="vertical" onFinish={handleChangePassword}>
                <Form.Item
                  name="old_password"
                  label={t('settings.currentPassword')}
                  rules={[{ required: true, message: t('settings.currentPasswordRequired') }]}
                >
                  <Input.Password prefix={<LockOutlined />} placeholder={t('settings.currentPasswordPlaceholder')} />
                </Form.Item>
                <Form.Item
                  name="new_password"
                  label={t('settings.newPassword')}
                  rules={[
                    { required: true, message: t('settings.newPasswordRequired') },
                    { min: 6, message: t('auth.passwordMin') },
                  ]}
                >
                  <Input.Password prefix={<LockOutlined />} placeholder={t('settings.newPasswordPlaceholder')} />
                </Form.Item>
                <Form.Item
                  name="confirm_password"
                  label={t('settings.confirmNewPassword')}
                  dependencies={['new_password']}
                  rules={[
                    { required: true, message: t('settings.confirmNewPasswordRequired') },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('new_password') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error(t('auth.passwordNotMatch')));
                      },
                    }),
                  ]}
                >
                  <Input.Password prefix={<LockOutlined />} placeholder={t('settings.confirmNewPasswordPlaceholder')} />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit" loading={pwLoading} danger>{t('settings.changePassword')}</Button>
                </Form.Item>
              </Form>
            </ContentCard>

            <ContentCard title={t('settings.language')}>
              <Form layout="vertical">
                <Form.Item label={t('settings.languageLabel')}>
                  <Select
                    value={i18n.language}
                    onChange={(lng) => i18n.changeLanguage(lng)}
                    options={[
                      { label: t('settings.chinese'), value: 'zh-CN' },
                      { label: t('settings.english'), value: 'en-US' },
                    ]}
                    style={{ width: 200 }}
                  />
                </Form.Item>
              </Form>
            </ContentCard>
          </div>
        </Col>
      </Row>
    </PageLayout>
  );
};

export default SettingsPage;
