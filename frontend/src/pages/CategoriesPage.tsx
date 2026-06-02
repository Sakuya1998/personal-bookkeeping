import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Tabs, Table, Button, Modal, Form, Input, InputNumber, Space, Popconfirm, message, Dropdown, Skeleton, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import client from '../api/client';
import { ApiResponse, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const CategoriesPage: React.FC = () => {
  const { t } = useTranslation();
  const { currentLedger } = useAppStore();
  const currentRole = useAppStore(s => s.currentRole);
  const canManage = currentRole === 'owner' || currentRole === 'admin';
  const [income, setIncome] = useState<Category[]>([]);
  const [expense, setExpense] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [form] = Form.useForm();
  const [type, setType] = useState<'income' | 'expense'>('expense');

  const load = useCallback(async () => {
    if (!currentLedger) return;
    setLoading(true);
    try {
      const res = await client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`);
      setIncome(res.data.data.filter((c: Category) => c.type === 'income'));
      setExpense(res.data.data.filter((c: Category) => c.type === 'expense'));
    } finally {
      setLoading(false);
    }
  }, [currentLedger]);

  useEffect(() => {
    if (!currentLedger) return;
    load().catch(err => console.error('获取分类失败:', err));
  }, [currentLedger, load]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      const data = {
        ...values,
        type,
        sort_order: values.sort_order,
      };
      if (editing) {
        await client.put(`/categories/${editing.id}`, data);
        message.success(t('categories.updateSuccess'));
      } else {
        await client.post('/categories', data);
        message.success(t('categories.createSuccess'));
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      load();
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } };
      message.error(e.response?.data?.message || t('common.failed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/categories/${id}`);
      message.success(t('common.success'));
      load();
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } };
      message.error(e.response?.data?.message || t('common.failed'));
    }
  };

  const openCreate = (nextType: 'income' | 'expense') => {
    setType(nextType);
    setEditing(null);
    form.resetFields();
    setModalOpen(true);
  };

  const openEdit = (cat: Category) => {
    setType(cat.type);
    setEditing(cat);
    form.setFieldsValue(cat);
    setModalOpen(true);
  };

  const columns = useMemo(() => [
    { title: t('categories.icon'), dataIndex: 'icon', key: 'icon', width: 60, render: (v: string) => v || '-' },
    { title: t('categories.name'), dataIndex: 'name', key: 'name' },
    { title: t('categories.sort'), dataIndex: 'sort_order', key: 'sort', width: 60 },
    {
      title: t('categories.action'), key: 'action', width: 100,
      render: (_: unknown, r: Category) => {
        if (!canManage) return null;
        return <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title={t('common.confirmDelete')} onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>;
      },
    },
  ], [t, canManage, openEdit, handleDelete]);

  const dataSource = useMemo(() => (type === 'expense' ? expense : income), [type, expense, income]);

  const createLabel = type === 'expense' ? t('categories.addExpense') : t('categories.addIncome');

  return (
    <PageLayout
      header={<PageTitle title={t('categories.title')} />}
      toolbar={(
        <PageToolbar
          left={(
            <Tabs
              activeKey={type}
              onChange={(k) => setType(k as 'income' | 'expense')}
              items={[
                { key: 'expense', label: t('categories.expenseTab'), children: null },
                { key: 'income', label: t('categories.incomeTab'), children: null },
              ]}
              tabBarStyle={{ margin: 0 }}
              style={{ margin: 0 }}
            />
          )}
          right={(
            <Dropdown.Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => openCreate(type)}
              menu={{
                items: [
                  { key: 'expense', label: t('categories.addExpense') },
                  { key: 'income', label: t('categories.addIncome') },
                ],
                onClick: ({ key }) => openCreate(key as 'income' | 'expense'),
              }}
            >
              {createLabel}
            </Dropdown.Button>
          )}
        />
      )}
    >
      <ContentCard>
        {loading && dataSource.length === 0 ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : dataSource.length === 0 ? (
          <Empty description={t('categories.empty')} />
        ) : (
          <Table dataSource={dataSource} columns={columns} rowKey="id" size="small" pagination={false} />
        )}
      </ContentCard>

      <Modal
        title={editing ? t('categories.edit') : createLabel}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label={t('categories.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="icon" label={t('categories.icon')}>
            <Input placeholder="🍽️" />
          </Form.Item>
          <Form.Item name="color" label={t('categories.color')}>
            <Input placeholder="#1890ff" />
          </Form.Item>
          <Form.Item name="sort_order" label={t('categories.sort')}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </PageLayout>
  );
};

export default CategoriesPage;
