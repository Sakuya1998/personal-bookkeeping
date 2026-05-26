import React, { useEffect, useState } from 'react';
import { Tabs, Table, Button, Modal, Form, Input, Space, Popconfirm, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import client from '../api/client';
import { ApiResponse, Category } from '../api/types';
import { useAppStore } from '../store/appStore';

const CategoriesPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [income, setIncome] = useState<Category[]>([]);
  const [expense, setExpense] = useState<Category[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [form] = Form.useForm();
  const [type, setType] = useState<'income' | 'expense'>('expense');

  const load = async () => {
    if (!currentLedger) return;
    const res = await client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`);
    setIncome(res.data.data.filter((c: Category) => c.type === 'income'));
    setExpense(res.data.data.filter((c: Category) => c.type === 'expense'));
  };

  useEffect(() => {
    if (currentLedger) {
      client.get<ApiResponse<Category[]>>(`/ledgers/${currentLedger.id}/categories`).then((res) => {
        setIncome(res.data.data.filter((c: Category) => c.type === 'income'));
        setExpense(res.data.data.filter((c: Category) => c.type === 'expense'));
      });
    }
  }, [currentLedger]);

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      const data = { ...values, type };
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

  const columns = [
    { title: '图标', dataIndex: 'icon', key: 'icon', width: 60, render: (v: string) => v || '-' },
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '排序', dataIndex: 'sort_order', key: 'sort', width: 60 },
    {
      title: '操作', key: 'action', width: 100,
      render: (_: unknown, r: Category) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => { setEditing(r); form.setFieldsValue(r); setModalOpen(true); }} />
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>分类管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setModalOpen(true); }}>新建分类</Button>
      </div>

      <Tabs activeKey={type} onChange={(k) => setType(k as 'income' | 'expense')}>
        <Tabs.TabPane tab="支出分类" key="expense">
          <Table dataSource={expense} columns={columns} rowKey="id" size="small" pagination={false} />
        </Tabs.TabPane>
        <Tabs.TabPane tab="收入分类" key="income">
          <Table dataSource={income} columns={columns} rowKey="id" size="small" pagination={false} />
        </Tabs.TabPane>
      </Tabs>

      <Modal
        title={editing ? '编辑分类' : '新建分类'}
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
            <Input type="number" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CategoriesPage;
