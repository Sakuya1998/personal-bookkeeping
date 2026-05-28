import React, { useEffect, useState, useCallback } from 'react';
import { Card, Row, Col, Button, Modal, Form, Input, Select, Tag, Popconfirm, message, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, WalletOutlined } from '@ant-design/icons';
import client from '../api/client';
import { ApiResponse, Ledger } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES } from '../utils/currency';

const LedgersPage: React.FC = () => {
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
      message.error(apiErr.response?.data?.message || '加载账本失败');
    }
  }, [setLedgers]);

  useEffect(() => {
    loadLedgers();
  }, [loadLedgers]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      if (editing) {
        await client.put(`/ledgers/${editing.id}`, values);
        message.success('更新成功');
      } else {
        await client.post('/ledgers', values);
        message.success('创建成功');
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadLedgers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '操作失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/ledgers/${id}`);
      message.success('删除成功');
      if (currentLedger?.id === id) {
        setCurrentLedger(null);
      }
      loadLedgers();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
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

  if (ledgers.length === 0) {
    return (
      <div>
        <Empty description="还没有账本，创建第一个吧">
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>创建账本</Button>
        </Empty>
        <Modal
          title="创建账本"
          open={modalOpen}
          onOk={form.submit}
          onCancel={() => { setModalOpen(false); setEditing(null); }}
          confirmLoading={loading}
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            <Form.Item name="name" label="名称" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="description" label="描述">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="base_currency" label="本位币">
              <Select options={CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code} - ${c.name}`, value: c.code }))} />
            </Form.Item>
            <Form.Item name="icon" label="图标">
              <Input placeholder="💰" />
            </Form.Item>
            <Form.Item name="color" label="颜色">
              <Input placeholder="#1890ff" />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>账本管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建账本</Button>
      </div>
      <Row gutter={[16, 16]}>
        {ledgers.map((ledger) => (
          <Col key={ledger.id} xs={24} sm={12} lg={8}>
            <Card
              hoverable
              onClick={() => setCurrentLedger(ledger)}
              style={currentLedger?.id === ledger.id ? { border: '2px solid #1890ff' } : {}}
              actions={[
                <EditOutlined key="edit" onClick={(e) => { e.stopPropagation(); openEdit(ledger); }} />,
                <Popconfirm key="del" title="确定删除？" onConfirm={(e) => { e?.stopPropagation(); handleDelete(ledger.id); }}>
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

      <Modal
        title={editing ? '编辑账本' : '创建账本'}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="base_currency" label="本位币">
            <Select options={CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code} - ${c.name}`, value: c.code }))} />
          </Form.Item>
          <Form.Item name="icon" label="图标">
            <Input placeholder="💰" />
          </Form.Item>
          <Form.Item name="color" label="颜色">
            <Input placeholder="#1890ff" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default LedgersPage;
