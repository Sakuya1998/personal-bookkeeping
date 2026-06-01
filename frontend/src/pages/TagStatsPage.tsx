import React, { useCallback, useEffect, useState } from 'react';
import { Table, DatePicker, Button, Skeleton, Empty, Statistic, Row, Col, message } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse } from '../api/types';
import { useAppStore } from '../store/appStore';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

interface TagStatsItem {
  tag: string;
  total_expense: number;
  total_income: number;
  transaction_count: number;
  percentage: number;
}

const TagStatsPage: React.FC = () => {
  const { t } = useTranslation();
  const currentLedger = useAppStore(s => s.currentLedger);
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<TagStatsItem[]>([]);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  const loadStats = useCallback(async (start?: string, end?: string) => {
    if (!currentLedger) return;
    setLoading(true);
    try {
      const params: Record<string, string> = {};
      if (start) params.start_date = start;
      if (end) params.end_date = end;
      const res = await client.get<ApiResponse<TagStatsItem[]>>(`/ledgers/${currentLedger.id}/tag-stats`, { params });
      setItems(res.data.data);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    } finally {
      setLoading(false);
    }
  }, [currentLedger, t]);

  useEffect(() => {
    if (currentLedger) loadStats();
  }, [currentLedger, loadStats]);

  const handleSearch = () => {
    const start = dateRange?.[0]?.format('YYYY-MM-DD');
    const end = dateRange?.[1]?.format('YYYY-MM-DD');
    loadStats(start, end);
  };

  const totalExpense = items.reduce((s, i) => s + i.total_expense, 0);
  const totalIncome = items.reduce((s, i) => s + i.total_income, 0);
  const totalTxn = items.reduce((s, i) => s + i.transaction_count, 0);

  const columns = [
    { title: t('tagStats.tag'), dataIndex: 'tag', key: 'tag', width: 160 },
    {
      title: <div style={{ textAlign: 'right' }}>{t('tagStats.expense')}</div>,
      dataIndex: 'total_expense', key: 'expense', align: 'right' as const, width: 160,
      render: (v: number) => v.toFixed(2),
    },
    {
      title: <div style={{ textAlign: 'right' }}>{t('tagStats.income')}</div>,
      dataIndex: 'total_income', key: 'income', align: 'right' as const, width: 160,
      render: (v: number) => v.toFixed(2),
    },
    {
      title: <div style={{ textAlign: 'right' }}>{t('tagStats.count')}</div>,
      dataIndex: 'transaction_count', key: 'count', align: 'right' as const, width: 100,
    },
    {
      title: <div style={{ textAlign: 'right' }}>{t('tagStats.percentage')}</div>,
      dataIndex: 'percentage', key: 'pct', align: 'right' as const, width: 100,
      render: (v: number) => `${v.toFixed(1)}%`,
    },
  ];

  return (
    <PageLayout
      header={<PageTitle title={t('tagStats.title')} />}
      toolbar={(
        <PageToolbar
          left={(
            <DatePicker.RangePicker
              style={{ width: 260 }}
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            />
          )}
          right={(
            <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
              {t('common.search')}
            </Button>
          )}
        />
      )}
    >
      <ContentCard>
        {items.length > 0 && (
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={6}><Statistic title={t('tagStats.totalExpense')} value={totalExpense} precision={2} prefix="¥" /></Col>
            <Col span={6}><Statistic title={t('tagStats.totalIncome')} value={totalIncome} precision={2} prefix="¥" /></Col>
            <Col span={6}><Statistic title={t('tagStats.transactionCount')} value={totalTxn} /></Col>
            <Col span={6}><Statistic title={t('tagStats.tagCount')} value={items.length} /></Col>
          </Row>
        )}
        {loading ? (
          <Skeleton active paragraph={{ rows: 8 }} />
        ) : items.length === 0 ? (
          <Empty description={dateRange ? t('tagStats.noTagsInRange') : t('tagStats.noTags')} />
        ) : (
          <Table dataSource={items} columns={columns} rowKey="tag" size="small" pagination={{ pageSize: 50 }} />
        )}
      </ContentCard>
    </PageLayout>
  );
};

export default TagStatsPage;
