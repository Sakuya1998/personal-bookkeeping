import React, { useEffect, useState, useCallback, useMemo } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, DatePicker,
  InputNumber, Tag, Space, message, Popconfirm, Switch, Row, Col, Skeleton, Empty, Divider,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SyncOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, RecurringRule, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES, formatCurrency } from '../utils/currency';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const RecurringPage: React.FC = () => {
  const { t } = useTranslation();
  const { currentLedger } = useAppStore();
  const [rules, setRules] = useState<RecurringRule[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<RecurringRule | null>(null);
  const [form] = Form.useForm();

  const frequency = Form.useWatch('frequency', form);

  const FREQ_OPTIONS = useMemo(() => [
    { label: t('recurring.daily'), value: 'daily' },
    { label: t('recurring.weekly'), value: 'weekly' },
    { label: t('recurring.monthly'), value: 'monthly' },
    { label: t('recurring.yearly'), value: 'yearly' },
  ], [t]);

  const WEEKDAY_OPTIONS = useMemo(() => [
    { label: t('calendar.sun'), value: 0 }, { label: t('calendar.mon'), value: 1 },
    { label: t('calendar.tue'), value: 2 }, { label: t('calendar.wed'), value: 3 },
    { label: t('calendar.thu'), value: 4 }, { label: t('calendar.fri'), value: 5 },
    { label: t('calendar.sat'), value: 6 },
  ], [t]);

  const FREQ_LABELS = useMemo<Record<string, string>>(() => ({
    daily: t('recurring.daily'),
    weekly: t('recurring.weekly'),
    monthly: t('recurring.monthly'),
    yearly: t('recurring.yearly'),
  }), [t]);

  const loadRules = useCallback(async () => {
    if (!currentLedger) return;
    queueMicrotask(() => setLoading(true));
    try {
      const res = await client.get<ApiResponse<RecurringRule[]>>('/recurring');
      setRules(res.data.data.filter((r) => r.ledger_id === currentLedger.id));
    } finally {
      setLoading(false);
    }
  }, [currentLedger]);

  useEffect(() => {
    if (!currentLedger) return;
    loadRules();
    client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`)
      .then((res) => setCategories(res.data.data))
      .catch(err => console.error(t('recurring.fetchCategoriesFailed'), err));
  }, [currentLedger, loadRules, t]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    const data = {
      ...values,
      ledger_id: currentLedger!.id,
      start_date: (values.start_date as dayjs.Dayjs).format('YYYY-MM-DD'),
      end_date: values.end_date ? (values.end_date as dayjs.Dayjs).format('YYYY-MM-DD') : null,
      tags: (values.tags as string[]) || [],
      day_of_month: values.day_of_month || null,
      weekday: values.weekday !== undefined ? values.weekday : null,
      interval: (values.interval as number) || 1,
    };

    try {
      if (editing) {
        await client.put(`/recurring/${editing.id}`, data);
        message.success(t('recurring.updateSuccess'));
      } else {
        await client.post('/recurring', data);
        message.success(t('recurring.createSuccess'));
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadRules();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/recurring/${id}`);
      message.success(t('recurring.deleteSuccess'));
      loadRules();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('recurring.deleteFailed'));
    }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ currency: 'CNY', frequency: 'monthly', interval: 1, start_date: dayjs() });
    setModalOpen(true);
  };

  const openEdit = (rule: RecurringRule) => {
    setEditing(rule);
    form.setFieldsValue({
      ...rule,
      start_date: dayjs(rule.start_date),
      end_date: rule.end_date ? dayjs(rule.end_date) : undefined,
      tags: rule.tags ? rule.tags.split(',') : [],
      interval: rule.interval || 1,
    });
    setModalOpen(true);
  };

  const columns = useMemo(() => [
    {
      title: t('transactions.type'), dataIndex: 'type', key: 'type', width: 70,
      render: (tVal: string) => (
        <Tag color={tVal === 'income' ? 'green' : 'red'}>
          {tVal === 'income' ? t('transactions.income') : t('transactions.expense')}
        </Tag>
      ),
    },
    {
      title: t('transactions.amount'), key: 'amount', width: 120,
      render: (_: unknown, r: RecurringRule) => (
        <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
          {r.type === 'income' ? '+' : '-'}{formatCurrency(r.amount, r.currency)}
        </span>
      ),
    },
    {
      title: t('recurring.frequency'), key: 'freq', width: 100,
      render: (_: unknown, r: RecurringRule) =>
        `${FREQ_LABELS[r.frequency] || r.frequency}${r.interval > 1 ? ` (${t('recurring.every')}${r.interval})` : ''}`,
    },
    {
      title: t('transactions.description'), dataIndex: 'description', key: 'desc',
      ellipsis: true, render: (v: string | null) => v || '-',
    },
    { title: t('recurring.startDate'), dataIndex: 'start_date', key: 'start', width: 110 },
    {
      title: t('recurring.endDate'), dataIndex: 'end_date', key: 'end', width: 110,
      render: (v: string | null) => v || t('recurring.noEndDate'),
    },
    { title: t('recurring.nextRun'), dataIndex: 'next_run_date', key: 'next', width: 110 },
    {
      title: t('recurring.status'), dataIndex: 'is_active', key: 'active', width: 80,
      render: (v: boolean) =>
        v ? <Tag color="green">{t('recurring.isActive')}</Tag> : <Tag color="default">{t('recurring.inactive')}</Tag>,
    },
    {
      title: t('categories.action'), key: 'action', width: 100,
      render: (_: unknown, r: RecurringRule) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title={t('recurring.deleteConfirm')} onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ], [t, FREQ_LABELS]);

  const catOptions = useMemo(
    () => categories.map((c) => ({ label: `${c.icon || ''} ${c.name}`, value: c.id })),
    [categories],
  );

  return (
    <PageLayout
      header={<PageTitle title={t('recurring.title')} description={t('recurring.pageDescription')} />}
      toolbar={(
        <PageToolbar
          left={(
            <Button icon={<SyncOutlined spin={loading} />} onClick={() => { loadRules(); }} disabled={loading}>
              {t('common.refresh')}
            </Button>
          )}
          right={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('recurring.add')}</Button>}
        />
      )}
    >
      <ContentCard>
        {loading && rules.length === 0 ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : rules.length === 0 ? (
          <Empty description={t('recurring.noRules')} />
        ) : (
          <Table
            dataSource={rules}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={false}
            size="small"
          />
        )}
      </ContentCard>

      <Modal
        title={editing ? t('recurring.edit') : t('recurring.add')}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        width={560}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Divider titlePlacement="left" style={{ marginTop: 0 }}>{t('recurring.basicInfo')}</Divider>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label={t('transactions.type')} rules={[{ required: true }]}>
                <Select options={[
                  { label: t('transactions.income'), value: 'income' },
                  { label: t('transactions.expense'), value: 'expense' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="amount" label={t('transactions.amount')} rules={[{ required: true }]}>
                <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} prefix="¥" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="category_id" label={t('transactions.category')} rules={[{ required: true }]}>
            <Select options={catOptions} />
          </Form.Item>

          <Form.Item name="currency" label={t('transactions.currency')}>
            <Select options={CURRENCIES.map((c) => ({ label: `${c.symbol} ${c.code}`, value: c.code }))} />
          </Form.Item>

          <Form.Item name="description" label={t('transactions.description')}>
            <Input.TextArea rows={2} />
          </Form.Item>

          <Divider titlePlacement="left">{t('recurring.ruleSection')}</Divider>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="frequency" label={t('recurring.frequency')} rules={[{ required: true }]}>
                <Select options={FREQ_OPTIONS} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="interval" label={t('recurring.interval')} tooltip={t('recurring.intervalTooltip')}>
                <InputNumber min={1} max={365} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          {frequency === 'weekly' && (
            <Form.Item name="weekday" label={t('recurring.weekday')}>
              <Select options={WEEKDAY_OPTIONS} />
            </Form.Item>
          )}

          {frequency === 'monthly' && (
            <Form.Item name="day_of_month" label={t('recurring.dayOfMonth')} tooltip={t('recurring.dayOfMonthTooltip')}>
              <InputNumber min={1} max={31} style={{ width: '100%' }} />
            </Form.Item>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="start_date" label={t('recurring.startDate')} rules={[{ required: true }]}>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="end_date" label={`${t('recurring.endDate')}（${t('recurring.optional')}）`}>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Divider titlePlacement="left">{t('recurring.advanced')}</Divider>
          <Form.Item name="tags" label={t('transactions.tags')}>
            <Select mode="tags" placeholder={t('transactions.tagPlaceholder')} />
          </Form.Item>

          {editing && (
            <Form.Item name="is_active" label={t('recurring.isActive')} valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </PageLayout>
  );
};

export default RecurringPage;
