import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Table, Tag, Empty, Button } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined, WalletOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import client from '../api/client';
import { ApiResponse, LedgerSummary, PaginatedData, Transaction } from '../api/types';
import { useAppStore } from '../store/appStore';
import { formatCurrency } from '../utils/currency';

const DashboardPage: React.FC = () => {
  const { currentLedger } = useAppStore();
  const [summary, setSummary] = useState<LedgerSummary | null>(null);
  const [recentTxns, setRecentTxns] = useState<Transaction[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    if (!currentLedger) return;
    client.get<ApiResponse<LedgerSummary>>(`/ledgers/${currentLedger.id}/summary`).then((res) => {
      setSummary(res.data.data);
    });
    client.get<ApiResponse<PaginatedData<Transaction>>>(`/ledgers/${currentLedger.id}/transactions?page=1&page_size=5`).then((res) => {
      setRecentTxns(res.data.data.items);
    });
  }, [currentLedger]);

  if (!currentLedger) {
    return (
      <Empty description="暂无账本">
        <Button type="primary" onClick={() => navigate('/ledgers')}>创建账本</Button>
      </Empty>
    );
  }

  const columns = [
    { title: '日期', dataIndex: 'transaction_date', key: 'date', width: 110 },
    {
      title: '分类', key: 'category', width: 120,
      render: (_, r: Transaction) => r.category?.name || '-',
    },
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 80,
      render: (t: string) => <Tag color={t === 'income' ? 'green' : 'red'}>{t === 'income' ? '收入' : '支出'}</Tag>,
    },
    {
      title: '金额', key: 'amount', width: 120,
      render: (_, r: Transaction) => (
        <span style={{ color: r.type === 'income' ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
          {r.type === 'income' ? '+' : '-'}{formatCurrency(r.base_amount, currentLedger.base_currency)}
        </span>
      ),
    },
    { title: '描述', dataIndex: 'description', key: 'desc', ellipsis: true },
  ];

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card hoverable onClick={() => navigate('/transactions?type=income')}>
            <Statistic
              title="总收入"
              value={summary?.total_income || 0}
              precision={2}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#52c41a' }}
              suffix={currentLedger.base_currency}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card hoverable onClick={() => navigate('/transactions?type=expense')}>
            <Statistic
              title="总支出"
              value={summary?.total_expense || 0}
              precision={2}
              prefix={<ArrowDownOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
              suffix={currentLedger.base_currency}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="结余"
              value={summary?.balance || 0}
              precision={2}
              prefix={<WalletOutlined />}
              valueStyle={{ color: (summary?.balance || 0) >= 0 ? '#1890ff' : '#ff4d4f' }}
              suffix={currentLedger.base_currency}
            />
          </Card>
        </Col>
      </Row>

      <Card title="最近记录" extra={<a onClick={() => navigate('/transactions')}>查看全部</a>}>
        <Table
          dataSource={recentTxns}
          columns={columns}
          rowKey="id"
          pagination={false}
          size="small"
        />
      </Card>

      {summary && summary.expense_by_category && summary.expense_by_category.length > 0 && (
        <Card title="支出分类排行" style={{ marginTop: 16 }}>
          {summary.expense_by_category.map((cat) => (
            <div key={cat.category_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
              <span>{cat.category_icon} {cat.category_name} ({cat.count}笔)</span>
              <span style={{ fontWeight: 600, color: '#ff4d4f' }}>{formatCurrency(cat.total, currentLedger.base_currency)}</span>
            </div>
          ))}
        </Card>
      )}
    </div>
  );
};

export default DashboardPage;
