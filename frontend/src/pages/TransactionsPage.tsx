import React, { useEffect, useState, useCallback, useRef } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, Select, DatePicker, Tag, Space, message, Popconfirm, Skeleton } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, TagsOutlined, CameraOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, PaginatedData, Transaction, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES, formatCurrency } from '../utils/currency';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';

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
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [batchCategoryModalOpen, setBatchCategoryModalOpen] = useState(false);
  const [batchCategoryId, setBatchCategoryId] = useState<string | undefined>(undefined);
  const [ocrLoading, setOcrLoading] = useState(false);

  const loadTxns = useCallback(async () => {
    if (!currentLedger) return;
    queueMicrotask(() => setLoading(true));
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

  useEffect(() => {
    if (!currentLedger) return;
    queueMicrotask(() => setSelectedRowKeys([]));
    loadTxns();
    client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`)
      .then((res) => setCategories(res.data.data))
      .catch(err => console.error('获取分类失败:', err));
  }, [currentLedger, page, pageSize, filters, loadTxns]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    const data = {
      ...values,
      amount: Number(values.amount),
      ledger_id: currentLedger!.id,
      transaction_date: (values.transaction_date as dayjs.Dayjs).format('YYYY-MM-DD'),
      tags: (values.tags as string[]) || [],
    };
    try {
      let overBudget = false;
      if (editing) {
        const res = await client.put<ApiResponse<{ transaction: Transaction; over_budget: boolean }>>(`/transactions/${editing.id}`, data);
        overBudget = res.data.data.over_budget;
        message.success('更新成功');
      } else {
        const res = await client.post<ApiResponse<{ transaction: Transaction; over_budget: boolean }>>('/transactions', data);
        overBudget = res.data.data.over_budget;
        message.success('创建成功');
      }
      if (overBudget) {
        message.warning({ content: '⚠️ 该笔交易已超出当月预算，请注意控制支出', duration: 5, key: 'budget_warning' });
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/transactions/${id}`);
      message.success('删除成功');
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
    }
  };

  const handleBatchDelete = () => {
    Modal.confirm({
      title: '批量删除',
      content: `确定要删除选中的 ${selectedRowKeys.length} 条记录吗？此操作不可恢复。`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await client.post<ApiResponse<{ deleted: number }>>('/transactions/batch-delete', {
            ids: selectedRowKeys,
          });
          message.success(`已删除 ${res.data.data.deleted} 条记录`);
          setSelectedRowKeys([]);
          loadTxns();
        } catch (err: unknown) {
          const apiErr = err as { response?: { data?: { message?: string } } };
          message.error(apiErr.response?.data?.message || '批量删除失败');
        }
      },
    });
  };

  const handleBatchCategorySubmit = async () => {
    if (!batchCategoryId || selectedRowKeys.length === 0) return;
    try {
      const res = await client.put<ApiResponse<{ updated: number }>>('/transactions/batch-update', {
        ids: selectedRowKeys,
        category_id: batchCategoryId,
      });
      message.success(`已更新 ${res.data.data.updated} 条记录的分类`);
      setBatchCategoryModalOpen(false);
      setBatchCategoryId(undefined);
      setSelectedRowKeys([]);
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '批量修改分类失败');
    }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ currency: 'CNY', transaction_date: dayjs(), tags: [] });
    setModalOpen(true);
  };

  const handleOCRUpload = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/jpeg,image/png';
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      setOcrLoading(true);
      try {
        const formData = new FormData();
        formData.append('image', file);
        const res = await client.post<ApiResponse<{ text: string; amount?: number; date?: string; merchant?: string }>>('/ocr/receipt', formData);
        const data = res.data.data;
        const vals: Record<string, unknown> = { currency: 'CNY', tags: [] };
        if (data.amount) vals.amount = data.amount;
        if (data.date) vals.transaction_date = dayjs(data.date);
        if (data.merchant) vals.description = data.merchant;
        if (data.text) vals.description = (vals.description ? vals.description + ' ' : '') + data.text.slice(0, 100);
        form.resetFields();
        form.setFieldsValue(vals);
        if (data.amount) message.success(`识别到金额: ¥${data.amount}${data.merchant ? '，商家: ' + data.merchant : ''}`);
        else message.info('未识别到金额，请手动填写');
        setEditing(null);
        setModalOpen(true);
      } catch (err: unknown) {
        const apiErr = err as { response?: { data?: { message?: string } } };
        message.error(apiErr.response?.data?.message || '识别失败');
      } finally {
        setOcrLoading(false);
      }
    };
    input.click();
  };

  const openEdit = (txn: Transaction) => {
    setEditing(txn);
    form.setFieldsValue({ ...txn, transaction_date: dayjs(txn.transaction_date), tags: txn.tags ? txn.tags.split(',') : [] });
    setModalOpen(true);
  };

  const catOptions = categories
    .filter(c => !filters.type || c.type === filters.type)
    .map(c => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }));

  const dateRangeValue: [dayjs.Dayjs, dayjs.Dayjs] | null = (filters.start_date && filters.end_date)
    ? [dayjs(filters.start_date), dayjs(filters.end_date)]
    : null;

  const columns = [
    { title: '日期', dataIndex: 'transaction_date', key: 'date', width: 110 },
    { title: '分类', key: 'category', width: 120, render: (_: unknown, r: Transaction) => {
      const cat = r.category;
      return cat ? `${cat.icon || ''} ${cat.name}` : '-';
    }},
    { title: '类型', dataIndex: 'type', key: 'type', width: 70,
      render: (t: string) => <Tag color={t === 'income' ? 'green' : 'red'}>{t === 'income' ? '收入' : '支出'}</Tag>,
    },
    { title: <div style={{ textAlign: 'right' }}>金额</div>, key: 'amount', width: 160,
      render: (_: unknown, r: Transaction) => {
        const cur = currentLedger?.base_currency || 'CNY';
        return (
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
              {r.type === 'income' ? '+' : '-'}{formatCurrency(r.base_amount, cur)}
              {r.currency !== cur && <Tag style={{ marginInlineStart: 6, marginInlineEnd: 0 }}>{r.currency} {r.amount}</Tag>}
            </span>
          </div>
        );
      },
    },
    { title: '描述', dataIndex: 'description', key: 'desc', ellipsis: true,
      render: (_: unknown, r: Transaction) => {
        const text = r.description || '';
        if (!filters.keyword || !text) return text || '-';
        const lower = text.toLowerCase();
        const kw = filters.keyword.toLowerCase();
        const idx = lower.indexOf(kw);
        if (idx === -1) return text;
        return (
          <span>
            {text.slice(0, idx)}
            <mark>{text.slice(idx, idx + kw.length)}</mark>
            {text.slice(idx + kw.length)}
          </span>
        );
      },
    },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: unknown, r: Transaction) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
  };

  return (
    <PageLayout
      header={<PageTitle title="交易记录" />}
      toolbar={(
        <PageToolbar
          left={(
            <>
              <Select
                allowClear
                placeholder="类型"
                style={{ width: 110 }}
                value={filters.type || undefined}
                options={[{ label: '收入', value: 'income' }, { label: '支出', value: 'expense' }]}
                onChange={(v) => setFilters(p => ({ ...p, type: v || '', category_id: '' }))}
              />
              <Select
                allowClear
                placeholder="分类"
                style={{ width: 160 }}
                value={filters.category_id || undefined}
                options={catOptions}
                onChange={(v) => setFilters(p => ({ ...p, category_id: v || '' }))}
              />
              <DatePicker.RangePicker
                style={{ width: 260 }}
                value={dateRangeValue}
                onChange={(dates) => setFilters(p => ({ ...p, start_date: dates?.[0]?.format('YYYY-MM-DD') || '', end_date: dates?.[1]?.format('YYYY-MM-DD') || '' }))}
              />
              <Input
                allowClear
                prefix={<SearchOutlined />}
                placeholder="搜索描述"
                style={{ width: 200 }}
                onChange={(e) => {
                  const value = e.target.value;
                  if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
                  searchTimeoutRef.current = setTimeout(() => {
                    setFilters(p => ({ ...p, keyword: value }));
                  }, 300);
                }}
              />
            </>
          )}
          right={(
            <>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增</Button>
              <Button icon={<CameraOutlined />} loading={ocrLoading} onClick={handleOCRUpload}>拍照记账</Button>
              {selectedRowKeys.length > 0 ? (
                <>
                  <Tag color="processing" style={{ marginInlineEnd: 0 }}>已选 {selectedRowKeys.length} 项</Tag>
                  <Button icon={<TagsOutlined />} onClick={() => {
                    setBatchCategoryId(undefined);
                    setBatchCategoryModalOpen(true);
                  }}>修改分类</Button>
                  <Button danger icon={<DeleteOutlined />} onClick={handleBatchDelete}>批量删除</Button>
                </>
              ) : null}
            </>
          )}
        />
      )}
    >
      {loading && txns.length === 0 ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <Table
          dataSource={txns}
          columns={columns}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          pagination={{ current: page, total, pageSize, onChange: (p) => setPage(p) }}
          size="small"
        />
      )}

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
            <InputNumber step={0.01} min={0.01} prefix="¥" style={{ width: '100%' }} />
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

      <Modal
        title="批量修改分类"
        open={batchCategoryModalOpen}
        onOk={handleBatchCategorySubmit}
        onCancel={() => { setBatchCategoryModalOpen(false); setBatchCategoryId(undefined); }}
        okText="确认修改"
        cancelText="取消"
        okButtonProps={{ disabled: !batchCategoryId }}
      >
        <p style={{ marginBottom: 16 }}>将为选中的 {selectedRowKeys.length} 条记录统一修改分类：</p>
        <Select
          placeholder="选择目标分类"
          style={{ width: '100%' }}
          value={batchCategoryId}
          onChange={(v) => setBatchCategoryId(v)}
          options={categories.map(c => ({ label: `${c.icon || ''} ${c.name} (${c.type === 'income' ? '收入' : '支出'})`, value: c.id }))}
        />
      </Modal>
    </PageLayout>
  );
};

export default TransactionsPage;
