import React, { useEffect, useState, useCallback } from 'react';
import {
  Card, Table, Button, Modal, Form, Input, Select, DatePicker,
  InputNumber, Tag, Space, message, Popconfirm, Switch, Row, Col, Skeleton, Empty,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SyncOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, RecurringRule, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import { CURRENCIES, formatCurrency } from '../utils/currency';

const FREQ_OPTIONS = [
  { label: '每天', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' },
  { label: '每年', value: 'yearly' },
];

const WEEKDAY_OPTIONS = [
  { label: '周日', value: 0 }, { label: '周一', value: 1 },
  { label: '周二', value: 2 }, { label: '周三', value: 3 },
  { label: '周四', value: 4 }, { label: '周五', value: 5 },
  { label: '周六', value: 6 },
];

const FREQ_LABELS: Record<string, string> = { daily: '每天', weekly: '每周', monthly: '每月', yearly: '每年' };

const RecurringPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [rules, setRules] = useState<RecurringRule[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<RecurringRule | null>(null);
  const [form] = Form.useForm();

  const frequency = Form.useWatch('frequency', form);

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
      .catch(err => console.error('获取分类失败:', err));
  }, [currentLedger, loadRules]);

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
        message.success('更新成功');
      } else {
        await client.post('/recurring', data);
        message.success('创建成功');
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadRules();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/recurring/${id}`);
      message.success('删除成功');
      loadRules();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
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

  const columns = [
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 70,
      render: (t: string) => <Tag color={t === 'income' ? 'green' : 'red'}>{t === 'income' ? '收入' : '支出'}</Tag>,
    },
    {
      title: '金额', key: 'amount', width: 120,
      render: (_: unknown, r: RecurringRule) => (
        <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
          {r.type === 'income' ? '+' : '-'}{formatCurrency(r.amount, r.currency)}
        </span>
      ),
    },
    { title: '频率', key: 'freq', width: 100,
      render: (_: unknown, r: RecurringRule) => `${FREQ_LABELS[r.frequency] || r.frequency}${r.interval > 1 ? ` (每${r.interval})` : ''}` },
    { title: '描述', dataIndex: 'description', key: 'desc', ellipsis: true, render: (v: string | null) => v || '-' },
    { title: '开始', dataIndex: 'start_date', key: 'start', width: 110 },
    { title: '结束', dataIndex: 'end_date', key: 'end', width: 110, render: (v: string | null) => v || '无' },
    { title: '下次执行', dataIndex: 'next_run_date', key: 'next', width: 110 },
    {
      title: '状态', dataIndex: 'is_active', key: 'active', width: 80,
      render: (v: boolean) => v ? <Tag color="green">启用</Tag> : <Tag color="default">停用</Tag>,
    },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: unknown, r: RecurringRule) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const catOptions = categories.map((c) => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }));

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <SyncOutlined spin={loading} />
              <span style={{ color: '#999' }}>周期性交易会自动在指定日期生成记账记录</span>
            </Space>
          </Col>
          <Col>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增规则</Button>
          </Col>
        </Row>
      </Card>

      {loading && rules.length === 0 ? (
        <Skeleton active paragraph={{ rows: 6 }} />
      ) : rules.length === 0 ? (
        <Empty description="暂无周期性规则" />
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

      <Modal
        title={editing ? '编辑周期性规则' : '新增周期性规则'}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        width={560}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={[{ label: '收入', value: 'income' }, { label: '支出', value: 'expense' }]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="amount" label="金额" rules={[{ required: true }]}>
                <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} prefix="¥" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="category_id" label="分类" rules={[{ required: true }]}>
            <Select options={catOptions} />
          </Form.Item>

          <Form.Item name="currency" label="币种">
            <Select options={CURRENCIES.map((c) => ({ label: `${c.symbol} ${c.code}`, value: c.code }))} />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="frequency" label="频率" rules={[{ required: true }]}>
                <Select options={FREQ_OPTIONS} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="interval" label="间隔" tooltip="每 N 个周期执行一次">
                <InputNumber min={1} max={365} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          {frequency === 'weekly' && (
            <Form.Item name="weekday" label="星期几">
              <Select options={WEEKDAY_OPTIONS} />
            </Form.Item>
          )}

          {frequency === 'monthly' && (
            <Form.Item name="day_of_month" label="每月几号" tooltip="1-31，如果超过当月天数则取最后一天">
              <InputNumber min={1} max={31} style={{ width: '100%' }} />
            </Form.Item>
          )}

          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>

          <Form.Item name="tags" label="标签">
            <Select mode="tags" placeholder="输入标签后回车" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="start_date" label="开始日期" rules={[{ required: true }]}>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="end_date" label="结束日期（可选）">
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          {editing && (
            <Form.Item name="is_active" label="启用" valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  );
};

export default RecurringPage;
