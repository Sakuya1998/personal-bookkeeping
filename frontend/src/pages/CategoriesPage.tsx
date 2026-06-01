import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Tabs, Table, Button, Modal, Form, Input, InputNumber, Space, Popconfirm, message, Dropdown, Skeleton, Empty } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import client from '../api/client';
import { ApiResponse, Category } from '../api/types';
import { useAppStore } from '../store/appStore';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const CategoriesPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [income, setIncome] = useState<Category[]>([]);
  const [expense, setExpense] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [form] = Form.useForm();
  const [type, setType] = useState<'income' | 'expense'>('expense');

  const load = useCallback(async () => {
    if (!currentLedger) return;
    queueMicrotask(() => setLoading(true));
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
        message.success('更新成功');
      } else {
        await client.post('/categories', data);
        message.success('创建成功');
      }
      setModalOpen(false);
      setEditing(null);
      form.resetFields();
      load();
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } };
      message.error(e.response?.data?.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/categories/${id}`);
      message.success('删除成功');
      load();
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } };
      message.error(e.response?.data?.message || '删除失败');
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

  const columns = [
    { title: '图标', dataIndex: 'icon', key: 'icon', width: 60, render: (v: string) => v || '-' },
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '排序', dataIndex: 'sort_order', key: 'sort', width: 60 },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: unknown, r: Category) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)} />
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const dataSource = useMemo(() => (type === 'expense' ? expense : income), [type, expense, income]);

  const createLabel = type === 'expense' ? '新建支出分类' : '新建收入分类';

  return (
    <PageLayout
      header={<PageTitle title="分类管理" />}
      toolbar={(
        <PageToolbar
          left={(
            <Tabs
              activeKey={type}
              onChange={(k) => setType(k as 'income' | 'expense')}
              items={[
                { key: 'expense', label: '支出分类', children: null },
                { key: 'income', label: '收入分类', children: null },
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
                  { key: 'expense', label: '新建支出分类' },
                  { key: 'income', label: '新建收入分类' },
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
          <Empty description="暂无分类" />
        ) : (
          <Table dataSource={dataSource} columns={columns} rowKey="id" size="small" pagination={false} />
        )}
      </ContentCard>

      <Modal
        title={editing ? '编辑分类' : createLabel}
        open={modalOpen}
        onOk={form.submit}
        onCancel={() => { setModalOpen(false); setEditing(null); }}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="icon" label="图标（Emoji）">
            <Input placeholder="🍽️" />
          </Form.Item>
          <Form.Item name="color" label="颜色">
            <Input placeholder="#1890ff" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </PageLayout>
  );
};

export default CategoriesPage;
