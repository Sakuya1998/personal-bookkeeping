import React, { useEffect, useState, useCallback, useRef } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, Select, DatePicker, Tag, Space, message, Popconfirm, Skeleton } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, TagsOutlined, CameraOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, PaginatedData, Transaction, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES, formatCurrency } from '../utils/currency';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import CurrencySelect from '../components/CurrencySelect';

const TransactionsPage: React.FC = () => {
  const { t } = useTranslation();
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
        message.success(t('transactions.updateSuccess'));
      } else {
        const res = await client.post<ApiResponse<{ transaction: Transaction; over_budget: boolean }>>('/transactions', data);
        overBudget = res.data.data.over_budget;
        message.success(t('transactions.createSuccess'));
      }
      if (overBudget) {
        message.warning({ content: t('transactions.budgetWarning'), duration: 5, key: 'budget_warning' });
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/transactions/${id}`);
      message.success(t('transactions.deleteSuccess'));
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('transactions.deleteFailed'));
    }
  };

  const handleBatchDelete = () => {
    Modal.confirm({
      title: t('transactions.batchDelete'),
      content: t('transactions.batchDeleteConfirm', { count: selectedRowKeys.length }),
      okText: t('common.confirmDelete'),
      okType: 'danger',
      cancelText: t('common.cancel'),
      onOk: async () => {
        try {
          const res = await client.post<ApiResponse<{ deleted: number }>>('/transactions/batch-delete', {
            ids: selectedRowKeys,
          });
          message.success(t('transactions.batchDeleted', { count: res.data.data.deleted }));
          setSelectedRowKeys([]);
          loadTxns();
        } catch (err: unknown) {
          const apiErr = err as { response?: { data?: { message?: string } } };
          message.error(apiErr.response?.data?.message || t('transactions.batchDeleteFailed'));
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
      message.success(t('transactions.batchCategoryUpdated', { count: res.data.data.updated }));
      setBatchCategoryModalOpen(false);
      setBatchCategoryId(undefined);
      setSelectedRowKeys([]);
      loadTxns();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('transactions.batchCategoryFailed'));
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
        if (data.amount) {
          message.success(data.merchant
            ? t('transactions.ocrSuccessWithMerchant', { amount: data.amount, merchant: data.merchant })
            : t('transactions.ocrSuccess', { amount: data.amount }));
        } else {
          message.info(t('transactions.ocrNoAmount'));
        }
        setEditing(null);
        setModalOpen(true);
      } catch (err: unknown) {
        const apiErr = err as { response?: { data?: { message?: string } } };
        message.error(apiErr.response?.data?.message || t('transactions.ocrFailed'));
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
    { title: t('transactions.date'), dataIndex: 'transaction_date', key: 'date', width: 110 },
    { title: t('transactions.category'), key: 'category', width: 120, render: (_: unknown, r: Transaction) => {
      const cat = r.category;
      return cat ? `${cat.icon || ''} ${cat.name}` : '-';
    }},
    { title: t('transactions.type'), dataIndex: 'type', key: 'type', width: 70,
      render: (tType: string) => <Tag color={tType === 'income' ? 'green' : 'red'}>{tType === 'income' ? t('transactions.income') : t('transactions.expense')}</Tag>,
    },
    { title: <div style={{ textAlign: 'right' }}>{t('transactions.amount')}</div>, key: 'amount', width: 160,
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
    { title: t('transactions.description'), dataIndex: 'description', key: 'desc', ellipsis: true,
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
      title: t('transactions.actions'), key: 'action', width: 100,
      render: (_: unknown, r: Transaction) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title={t('transactions.deleteConfirm')} onConfirm={() => handleDelete(r.id)}>
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
      header={<PageTitle title={t('transactions.title')} />}
      toolbar={(
        <PageToolbar
          left={(
            <>
              <Select
                allowClear
                placeholder={t('transactions.type')}
                style={{ width: 110 }}
                value={filters.type || undefined}
                options={[{ label: t('transactions.income'), value: 'income' }, { label: t('transactions.expense'), value: 'expense' }]}
                onChange={(v) => setFilters(p => ({ ...p, type: v || '', category_id: '' }))}
              />
              <Select
                allowClear
                placeholder={t('transactions.category')}
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
                placeholder={t('transactions.searchPlaceholder')}
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
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('transactions.add')}</Button>
              <Button icon={<CameraOutlined />} loading={ocrLoading} onClick={handleOCRUpload}>{t('transactions.ocr')}</Button>
              {selectedRowKeys.length > 0 ? (
                <>
                  <Tag color="processing" style={{ marginInlineEnd: 0 }}>{t('transactions.selected')} {selectedRowKeys.length} {t('transactions.items')}</Tag>
                  <Button icon={<TagsOutlined />} onClick={() => {
                    setBatchCategoryId(undefined);
                    setBatchCategoryModalOpen(true);
                  }}>{t('transactions.batchCategory')}</Button>
                  <Button danger icon={<DeleteOutlined />} onClick={handleBatchDelete}>{t('transactions.batchDelete')}</Button>
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
        title={editing ? t('transactions.edit') : t('transactions.add')}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        width={500}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="type" label={t('transactions.type')} rules={[{ required: true }]}>
            <Select options={[{ label: t('transactions.income'), value: 'income' }, { label: t('transactions.expense'), value: 'expense' }]} onChange={() => form.setFieldValue('category_id', undefined)} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.type !== cur.type}>
            {({ getFieldValue }) => {
              const type = getFieldValue('type');
              const filtered = categories.filter(c => c.type === type);
              return (
                <Form.Item name="category_id" label={t('transactions.category')} rules={[{ required: true }]}>
                  <Select options={filtered.map(c => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }))} />
                </Form.Item>
              );
            }}
          </Form.Item>
          <Form.Item name="amount" label={t('transactions.amount')} rules={[{ required: true }]}>
            <InputNumber step={0.01} min={0.01} prefix="¥" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="currency" label={t('transactions.currency')}>
            <CurrencySelect style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label={t('transactions.description')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="transaction_date" label={t('transactions.date')} rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="tags" label={t('transactions.tags')}>
            <Select mode="tags" placeholder={t('transactions.tagPlaceholder')} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('transactions.batchCategory')}
        open={batchCategoryModalOpen}
        onOk={handleBatchCategorySubmit}
        onCancel={() => { setBatchCategoryModalOpen(false); setBatchCategoryId(undefined); }}
        okText={t('transactions.confirmModify')}
        cancelText={t('common.cancel')}
        okButtonProps={{ disabled: !batchCategoryId }}
      >
        <p style={{ marginBottom: 16 }}>{t('transactions.batchCategoryDesc', { count: selectedRowKeys.length })}</p>
        <Select
          placeholder={t('transactions.selectCategory')}
          style={{ width: '100%' }}
          value={batchCategoryId}
          onChange={(v) => setBatchCategoryId(v)}
          options={categories.map(c => ({ label: `${c.icon || ''} ${c.name} (${c.type === 'income' ? t('transactions.income') : t('transactions.expense')})`, value: c.id }))}
        />
      </Modal>
    </PageLayout>
  );
};

export default TransactionsPage;
