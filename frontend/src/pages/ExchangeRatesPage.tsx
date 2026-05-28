import React, { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, DatePicker, Select, Popconfirm, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, ExchangeRate } from '../api/types';
import { CURRENCIES } from '../utils/currency';

const ExchangeRatesPage: React.FC = () => {
  const [rates, setRates] = useState<ExchangeRate[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    client.get<ApiResponse<ExchangeRate[]>>('/exchange-rates').then((res) => {
      setRates(res.data.data);
    }).catch(err => console.error('获取汇率失败:', err));
  }, []);

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      await client.post('/exchange-rates', {
        ...values,
        date: (values.date as dayjs.Dayjs).format('YYYY-MM-DD'),
      });
      message.success('保存成功');
      setModalOpen(false);
      form.resetFields();
      client.get<ApiResponse<ExchangeRate[]>>('/exchange-rates').then((res) => {
        setRates(res.data.data);
      });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/exchange-rates/${id}`);
      message.success('删除成功');
      client.get<ApiResponse<ExchangeRate[]>>('/exchange-rates').then((res) => {
        setRates(res.data.data);
      });
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
    }
  };

  const currencyOpts = CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code}`, value: c.code }));

  const columns = [
    { title: '源币种', dataIndex: 'from_currency', key: 'from' },
    { title: '目标币种', dataIndex: 'to_currency', key: 'to' },
    { title: '汇率', dataIndex: 'rate', key: 'rate', render: (v: number) => v.toFixed(6) },
    { title: '日期', dataIndex: 'date', key: 'date' },
    { title: '来源', dataIndex: 'source', key: 'source' },
    {
      title: '操作', key: 'action', width: 80,
      render: (_: unknown, r: ExchangeRate) => (
        <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>汇率管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); form.setFieldsValue({ date: dayjs() }); setModalOpen(true); }}>新增汇率</Button>
      </div>

      <Table dataSource={rates} columns={columns} rowKey="id" size="small" pagination={{ pageSize: 50 }} />

      <Modal title="新增汇率" open={modalOpen} onOk={form.submit} onCancel={() => setModalOpen(false)}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="from_currency" label="源币种" rules={[{ required: true }]}>
            <Select options={currencyOpts} />
          </Form.Item>
          <Form.Item name="to_currency" label="目标币种" rules={[{ required: true }]}>
            <Select options={currencyOpts} />
          </Form.Item>
          <Form.Item name="rate" label="汇率" rules={[{ required: true }]}>
            <InputNumber step="0.000001" min="0.000001" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="date" label="日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="source" label="来源">
            <Input placeholder="例如：手动录入" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ExchangeRatesPage;
