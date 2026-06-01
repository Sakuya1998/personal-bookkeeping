import React, { useEffect, useState, useCallback } from 'react';
import { Card, Row, Col, Button, Modal, Form, Input, Select, Tag, Popconfirm, message, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, WalletOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import client from '../api/client';
import { ApiResponse, Ledger } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES } from '../utils/currency';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const LedgersPage: React.FC = () => {
  const { t } = useTranslation();
  const { ledgers, setLedgers, setCurrentLedger, currentLedger } = useAppStore();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Ledger | null>(null);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const loadLedgers = useCallback(async () => {
    try {
      const res = await client.get<ApiResponse<Ledger[]>>('/ledgers');
      setLedgers(res.data.data);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  }, [setLedgers, t]);

  useEffect(() => {
    loadLedgers();
  }, [loadLedgers]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      if (editing) {
        await client.put(`/ledgers/${editing.id}`, values);
        message.success(t('ledgers.updateSuccess'));
      } else {
        await client.post('/ledgers', values);
        message.success(t('ledgers.createSuccess'));
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadLedgers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/ledgers/${id}`);
      message.success(t('common.success'));
      if (currentLedger?.id === id) {
        setCurrentLedger(null);
      }
      loadLedgers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const openEdit = (ledger: Ledger) => {
    setEditing(ledger);
    form.setFieldsValue(ledger);
    setModalOpen(true);
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ base_currency: 'CNY' });
    setModalOpen(true);
  };

  return (
    <PageLayout
      header={<PageTitle title={t('nav.ledgers')} />}
      toolbar={(
        <PageToolbar
          right={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('ledgers.add')}</Button>}
        />
      )}
    >
      <ContentCard>
        {ledgers.length === 0 ? (
          <Empty description={t('dashboard.noLedger')}>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('ledgers.add')}</Button>
          </Empty>
        ) : (
          <Row gutter={[16, 16]}>
            {ledgers.map((ledger) => (
              <Col key={ledger.id} xs={24} sm={12} lg={8}>
                <Card
                  hoverable
                  onClick={() => setCurrentLedger(ledger)}
                  style={currentLedger?.id === ledger.id ? { border: '2px solid #1890ff' } : {}}
                  actions={[
                    <EditOutlined key="edit" onClick={(e) => { e.stopPropagation(); openEdit(ledger); }} />,
                    <Popconfirm key="del" title={t('ledgers.deleteConfirm')} onConfirm={(e) => { e?.stopPropagation(); handleDelete(ledger.id); }}>
                      <DeleteOutlined onClick={(e) => e.stopPropagation()} />
                    </Popconfirm>,
                  ]}
                >
                  <Card.Meta
                    avatar={<WalletOutlined style={{ fontSize: 28, color: ledger.color || '#1890ff' }} />}
                    title={ledger.name}
                    description={
                      <div>
                        <div>{ledger.description}</div>
                        <Tag color="blue" style={{ marginTop: 8 }}>{ledger.base_currency}</Tag>
                      </div>
                    }
                  />
                </Card>
              </Col>
            ))}
          </Row>
        )}
      </ContentCard>

      <Modal
        title={editing ? t('ledgers.edit') : t('ledgers.add')}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label={t('ledgers.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label={t('ledgers.description')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="base_currency" label={t('ledgers.baseCurrency')}>
            <Select options={CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code} - ${c.name}`, value: c.code }))} />
          </Form.Item>
          <Form.Item name="icon" label={t('ledgers.icon')}>
            <Input placeholder="💰" />
          </Form.Item>
          <Form.Item name="color" label={t('ledgers.color')}>
            <Input placeholder="#1890ff" />
          </Form.Item>
        </Form>
      </Modal>
    </PageLayout>
  );
};

export default LedgersPage;
