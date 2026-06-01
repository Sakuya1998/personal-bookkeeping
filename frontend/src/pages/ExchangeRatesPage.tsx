import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Table, Button, Popconfirm, Select, DatePicker, Input, message, Skeleton, Empty } from 'antd';
import { SyncOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, ExchangeRate } from '../api/types';
import { CURRENCIES } from '../utils/currency';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const ExchangeRatesPage: React.FC = () => {
  const [rates, setRates] = useState<ExchangeRate[]>([]);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);

  const [filters, setFilters] = useState<{
    from_currency: string;
    to_currency: string;
    source: string;
    dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null;
  }>({ from_currency: '', to_currency: '', source: '', dateRange: null });

  const loadRates = useCallback(async () => {
    queueMicrotask(() => setLoading(true));
    try {
      const res = await client.get<ApiResponse<ExchangeRate[]>>('/exchange-rates');
      setRates(res.data.data);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRates().catch(err => console.error('获取汇率失败:', err));
  }, [loadRates]);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await client.post('/exchange-rates/sync');
      message.success('汇率同步成功');
      loadRates();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '同步失败');
    } finally {
      setSyncing(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/exchange-rates/${id}`);
      message.success('删除成功');
      loadRates();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || '删除失败');
    }
  };

  const currencyOpts = CURRENCIES.map(c => ({ label: `${c.symbol} ${c.code}`, value: c.code }));

  const filteredRates = useMemo(() => {
    const kw = filters.source.trim().toLowerCase();
    return rates.filter((r) => {
      if (filters.from_currency && r.from_currency !== filters.from_currency) return false;
      if (filters.to_currency && r.to_currency !== filters.to_currency) return false;
      if (kw && !(r.source || '').toLowerCase().includes(kw)) return false;
      if (filters.dateRange) {
        const d = dayjs(r.date);
        if (d.isBefore(filters.dateRange[0], 'day')) return false;
        if (d.isAfter(filters.dateRange[1], 'day')) return false;
      }
      return true;
    });
  }, [rates, filters]);

  const hasFilters = Boolean(filters.from_currency || filters.to_currency || filters.source.trim() || filters.dateRange);

  const columns = [
    { title: '源币种', dataIndex: 'from_currency', key: 'from' },
    { title: '目标币种', dataIndex: 'to_currency', key: 'to' },
    { title: <div style={{ textAlign: 'right' }}>汇率</div>, dataIndex: 'rate', key: 'rate', align: 'right' as const, width: 140, render: (v: number) => v.toFixed(6) },
    { title: <div style={{ textAlign: 'right' }}>日期</div>, dataIndex: 'date', key: 'date', align: 'right' as const, width: 120 },
    { title: '来源', dataIndex: 'source', key: 'source' },
    {
      title: '操作', key: 'action', width: 80,
      render: (_: unknown, r: ExchangeRate) => (
        <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <PageLayout
      header={<PageTitle title="汇率管理" />}
      toolbar={(
        <PageToolbar
          left={(
            <>
              <Select
                allowClear
                placeholder="源币种"
                style={{ width: 140 }}
                value={filters.from_currency || undefined}
                options={currencyOpts}
                onChange={(v) => setFilters(p => ({ ...p, from_currency: v || '' }))}
              />
              <Select
                allowClear
                placeholder="目标币种"
                style={{ width: 140 }}
                value={filters.to_currency || undefined}
                options={currencyOpts}
                onChange={(v) => setFilters(p => ({ ...p, to_currency: v || '' }))}
              />
              <DatePicker.RangePicker
                style={{ width: 260 }}
                value={filters.dateRange}
                onChange={(dates) => setFilters(p => ({ ...p, dateRange: (dates as [dayjs.Dayjs, dayjs.Dayjs] | null) }))}
              />
              <Input
                allowClear
                placeholder="来源"
                style={{ width: 200 }}
                value={filters.source}
                onChange={(e) => setFilters(p => ({ ...p, source: e.target.value }))}
              />
            </>
          )}
          right={(
            <Button
              type="primary"
              icon={<SyncOutlined />}
              loading={syncing}
              onClick={handleSync}
            >
              手动同步
            </Button>
          )}
        />
      )}
    >
      <ContentCard>
        {loading && rates.length === 0 ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : filteredRates.length === 0 ? (
          <Empty description={hasFilters ? '无匹配记录' : '暂无汇率记录'} />
        ) : (
          <Table dataSource={filteredRates} columns={columns} rowKey="id" size="small" pagination={{ pageSize: 50 }} />
        )}
      </ContentCard>
    </PageLayout>
  );
};

export default ExchangeRatesPage;
