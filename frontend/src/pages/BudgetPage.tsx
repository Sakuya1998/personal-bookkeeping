import React, { useEffect, useState, useCallback } from 'react';
import {
  Card, Table, Button, Modal, Form, Select, InputNumber,
  message, Popconfirm, Space, Row, Col, Progress, Skeleton, Empty, Tag, DatePicker,
} from 'antd';
import { PlusOutlined, DeleteOutlined, FundOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, Budget, BudgetStatusItem, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import { formatCurrency } from '../utils/currency';

const BudgetPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [status, setStatus] = useState<BudgetStatusItem[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [month, setMonth] = useState(dayjs().format('YYYY-MM'));
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Budget | null>(null);
  const [form] = Form.useForm();

  const loadData = useCallback(async () => {
    if (!currentLedger) return;
    queueMicrotask(() => setLoading(true));
    try {
      const [budRes, statRes] = await Promise.all([
        client.get<ApiResponse<Budget[]>>(`/budgets?month=${month}`),
        client.get<ApiResponse<BudgetStatusItem[]>>(`/budgets/status?month=${month}&ledger_id=${currentLedger.id}`),
      ]);
      setBudgets((budRes.data.data || []).filter((b) => b.ledger_id === currentLedger.id));
      setStatus(statRes.data.data || []);
    } finally {
      setLoading(false);
    }
  }, [currentLedger, month]);

  useEffect(() => {
    if (!currentLedger) return;
    loadData();
    client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`)
      .then((res) => setCategories(res.data.data))
      .catch(err => console.error('获取分类失败:', err));
  }, [currentLedger, loadData]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      await client.post('/budgets', {
        ledger_id: currentLedger!.id,
        category_id: values.category_id || null,
        month,
        amount: values.amount,
      });
      message.success(editing ? '预算已更新' : '预算已创建');
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      loadData();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/budgets/${id}`);
      message.success('删除成功');
      loadData();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
    }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setModalOpen(true);
  };

  const expenseCategories = categories.filter((c) => c.type === 'expense');

  const budgetColumns = [
    {
      title: '分类', key: 'category', width: 150,
      render: (_: unknown, r: Budget) => {
        if (!r.category_id) return <Tag>全部支出</Tag>;
        const cat = categories.find((c) => c.id === r.category_id);
        return cat ? `${cat.icon || ''} ${cat.name}` : r.category_id;
      },
    },
    {
      title: '预算金额', key: 'amount', width: 150,
      render: (_: unknown, r: Budget) => (
        <span style={{ fontWeight: 600 }}>{formatCurrency(r.amount, 'CNY')}</span>
      ),
    },
    {
      title: '月份', dataIndex: 'month', key: 'month', width: 100,
    },
    {
      title: '操作', key: 'action', width: 80,
      render: (_: unknown, r: Budget) => (
        <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const formatPct = (pct: number) => `${Math.min(pct, 999).toFixed(1)}%`;

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <DatePicker
                picker="month"
                value={dayjs(month)}
                onChange={(d) => d && setMonth(d.format('YYYY-MM'))}
                allowClear={false}
              />
            </Space>
          </Col>
          <Col>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增预算</Button>
          </Col>
        </Row>
      </Card>

      <Row gutter={16}>
        {/* Budget status */}
        <Col xs={24} lg={14} style={{ marginBottom: 16 }}>
          <Card title={<><FundOutlined /> 预算执行状态</>}>
            {loading && status.length === 0 ? (
              <Skeleton active paragraph={{ rows: 4 }} />
            ) : status.length === 0 ? (
              <Empty description="本月暂无预算" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              status.map((s, idx) => {
                const over80 = s.percentage >= 80;
                const over100 = s.percentage >= 100;
                const color = over100 ? '#ff4d4f' : over80 ? '#faad14' : '#52c41a';
                return (
                  <div key={s.budget_id || idx} style={{ marginBottom: 16 }}>
                    <Row justify="space-between" style={{ marginBottom: 4 }}>
                      <Col>
                        <span>
                          {s.icon} {s.name || '全部支出'}
                        </span>
                      </Col>
                      <Col>
                        <span style={{ color }}>
                          {formatCurrency(s.spent, 'CNY')} / {formatCurrency(s.budget, 'CNY')}
                        </span>
                        <span style={{ marginLeft: 8, fontWeight: 600, color }}>
                          {formatPct(s.percentage)}
                        </span>
                      </Col>
                    </Row>
                    <Progress
                      percent={s.percentage}
                      strokeColor={color}
                      showInfo={false}
                      size="small"
                    />
                  </div>
                );
              })
            )}
          </Card>
        </Col>

        {/* Budget list */}
        <Col xs={24} lg={10} style={{ marginBottom: 16 }}>
          <Card title="预算设置">
            {loading && budgets.length === 0 ? (
              <Skeleton active paragraph={{ rows: 3 }} />
            ) : budgets.length === 0 ? (
              <Empty description="暂无预算设置" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <Table
                dataSource={budgets}
                columns={budgetColumns}
                rowKey="id"
                loading={loading}
                pagination={false}
                size="small"
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Create/Edit Modal */}
      <Modal
        title="新增预算"
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
        width={400}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="category_id" label="分类（不选则为全局预算）">
            <Select
              allowClear
              placeholder="留空 = 全部支出"
              options={expenseCategories.map((c) => ({ label: `${c.icon || ''} ${c.name}`, value: c.id }))}
            />
          </Form.Item>
          <Form.Item name="amount" label="预算金额" rules={[{ required: true, message: '请输入预算金额' }]}>
            <InputNumber min={0.01} step={0.01} style={{ width: '100%' }} prefix="¥" placeholder="例如 5000" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default BudgetPage;
