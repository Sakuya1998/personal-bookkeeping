import React, { useEffect, useState, useCallback } from 'react';
import { Card, Table, Button, Modal, Form, Input, Select, DatePicker, Tag, Space, message, Popconfirm, Row, Col } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, PaginatedData, Transaction, Category, Ledger } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES, formatCurrency } from '../utils/currency';

const TransactionsPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [txns, setTxns] = useState<Transaction[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState({ type: '', category_id: '', keyword: '', start_date: '', end_date: '' });

  const loadTxns = useCallback(async () => {
    if (!currentLedger) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) });
      Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v); });
      const res = await client.get<ApiResponse<PaginatedData<Transaction>>>(`/ledgers/${currentLedger.id}/transactions?${params}`);
      setTxns(res.data.data.items);
      setTotal(res.data.data.total);
    } finally {
      setLoading(false);
    }
  }, [currentLedger, page, pageSize, filters]);

  const loadCategories = async () => {
    if (!currentLedger) return;
    const res = await client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`);
    setCategories(res.data.data);
  };

  useEffect(() => { loadTxns(); loadCategories(); }, [currentLedger, page]);

  const handleSubmit = async (values: any) => {
    const data = {
      ...values,
      ledger_id: currentLedger!.id,
      transaction_date: values.transaction_date.format('YYYY-MM-DD'),
      tags: values.tags || [],
    };
    try {
      if (editing) {
        await client.put(`/transactions/${editing.id}`, data);
        message.success('更新成功');
      } else {
        await client.post('/transactions', data);
        message.success('创建成功');
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadTxns();
    } catch (err: any) {
      message.error(err.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    await client.delete(`/transactions/${id}`);
    message.success('删除成功');
    loadTxns();
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ currency: 'CNY', transaction_date: dayjs(), tags: [] });
    setModalOpen(true);
  };

  const openEdit = (txn: Transaction) => {
    setEditing(txn);
    form.setFieldsValue({ ...txn, transaction_date: dayjs(txn.transaction_date), tags: txn.tags ? txn.tags.split(',') : [] });
    setModalOpen(true);
  };

  const catOptions = categories
    .filter(c => !filters.type || c.type === filters.type)
    .map(c => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }));

  const columns = [
    { title: '日期', dataIndex: 'transaction_date', key: 'date', width: 110 },
    { title: '分类', key: 'category', width: 120, render: (_: any, r: Transaction) => {
      const cat = r.category;
      return cat ? `${cat.icon || ''} ${cat.name}` : '-';
    }},
    { title: '类型', dataIndex: 'type', key: 'type', width: 70,
      render: (t: string) => <Tag color={t === 'income' ? 'green' : 'red'}>{t === 'income' ? '收入' : '支出'}</Tag>,
    },
    { title: '金额', key: 'amount', width: 150,
      render: (_: any, r: Transaction) => {
        const cur = currentLedger?.base_currency || 'CNY';
        return (
          <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
            {r.type === 'income' ? '+' : '-'}{formatCurrency(r.base_amount, cur)}
            {r.currency !== cur && <Tag style={{ marginLeft: 4 }}>{r.currency} {r.amount}</Tag>}
          </span>
        );
      },
    },
    { title: '描述', dataIndex: 'description', key: 'desc', ellipsis: true },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: any, r: Transaction) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[8, 8]} align="middle">
          <Col><Select allowClear placeholder="类型" style={{ width: 100 }} options={[{ label: '收入', value: 'income' }, { label: '支出', value: 'expense' }]} onChange={(v) => setFilters(p => ({ ...p, type: v || '', category_id: '' }))} /></Col>
          <Col><Select allowClear placeholder="分类" style={{ width: 140 }} options={catOptions} onChange={(v) => setFilters(p => ({ ...p, category_id: v || '' }))} /></Col>
          <Col><DatePicker.RangePicker onChange={(dates) => setFilters(p => ({ ...p, start_date: dates?.[0]?.format('YYYY-MM-DD') || '', end_date: dates?.[1]?.format('YYYY-MM-DD') || '' }))} /></Col>
          <Col><Input prefix={<SearchOutlined />} placeholder="搜索描述" style={{ width: 160 }} onChange={(e) => setFilters(p => ({ ...p, keyword: e.target.value }))} /></Col>
          <Col>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增</Button>
          </Col>
        </Row>
      </Card>

      <Table
        dataSource={txns}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{ current: page, total, pageSize, onChange: (p) => setPage(p) }}
        size="small"
      />

      <Modal
        title={editing ? '编辑记录' : '新增记录'}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        width={500}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ label: '收入', value: 'income' }, { label: '支出', value: 'expense' }]} onChange={() => form.setFieldValue('category_id', undefined)} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.type !== cur.type}>
            {({ getFieldValue }) => {
              const type = getFieldValue('type');
              const filtered = categories.filter(c => c.type === type);
              return (
                <Form.Item name="category_id" label="分类" rules={[{ required: true }]}>
                  <Select options={filtered.map(c => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }))} />
                </Form.Item>
              );
            }}
          </Form.Item>
          <Form.Item name="amount" label="金额" rules={[{ required: true }]}>
            <Input type="number" step="0.01" min="0.01" prefix="¥" />
          </Form.Item>
          <Form.Item name="currency" label="币种">
            <Select options={CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code}`, value: c.code }))} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="transaction_date" label="日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="tags" label="标签">
            <Select mode="tags" placeholder="输入标签后回车" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TransactionsPage;
