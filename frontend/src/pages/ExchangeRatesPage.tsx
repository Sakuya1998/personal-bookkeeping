import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Table, Button, Popconfirm, Select, DatePicker, Input, message, Skeleton, Empty } from 'antd';
import { SyncOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, ExchangeRate } from '../api/types';
import CurrencySelect from '../components/CurrencySelect';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const ExchangeRatesPage: React.FC = () => {
  const { t } = useTranslation();
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
    setLoading(true);
    try {
      const res = await client.get<ApiResponse<ExchangeRate[]>>('/exchange-rates');
      setRates(res.data.data);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRates().catch(err => { console.error('获取汇率失败:', err); message.error(t('common.failed')); });
  }, [loadRates]);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await client.post('/exchange-rates/sync');
      message.success(t('exchangeRates.syncSuccess'));
      loadRates();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('exchangeRates.syncFailed'));
    } finally {
      setSyncing(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await client.delete(`/exchange-rates/${id}`);
      message.success(t('common.success'));
      loadRates();
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.failed'));
    }
  };

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
    { title: t('exchangeRates.fromCurrency'), dataIndex: 'from_currency', key: 'from' },
    { title: t('exchangeRates.toCurrency'), dataIndex: 'to_currency', key: 'to' },
    { title: <div style={{ textAlign: 'right' }}>{t('exchangeRates.rate')}</div>, dataIndex: 'rate', key: 'rate', align: 'right' as const, width: 140, render: (v: number) => v.toFixed(6) },
    { title: <div style={{ textAlign: 'right' }}>{t('exchangeRates.date')}</div>, dataIndex: 'date', key: 'date', align: 'right' as const, width: 120 },
    { title: t('exchangeRates.source'), dataIndex: 'source', key: 'source' },
    {
      title: t('exchangeRates.action'), key: 'action', width: 80,
      render: (_: unknown, r: ExchangeRate) => (
        <Popconfirm title={t('exchangeRates.deleteConfirm')} onConfirm={() => handleDelete(r.id)}>
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <PageLayout
      header={<PageTitle title={t('exchangeRates.title')} />}
      toolbar={(
        <PageToolbar
          left={(
            <>
              <CurrencySelect
                allowClear
                placeholder={t('exchangeRates.fromCurrency')}
                style={{ width: 140 }}
                value={filters.from_currency || undefined}
                onChange={(v) => setFilters(p => ({ ...p, from_currency: v || '' }))}
              />
              <CurrencySelect
                allowClear
                placeholder={t('exchangeRates.toCurrency')}
                style={{ width: 140 }}
                value={filters.to_currency || undefined}
                onChange={(v) => setFilters(p => ({ ...p, to_currency: v || '' }))}
              />
              <DatePicker.RangePicker
                style={{ width: 260 }}
                value={filters.dateRange}
                onChange={(dates) => setFilters(p => ({ ...p, dateRange: (dates as [dayjs.Dayjs, dayjs.Dayjs] | null) }))}
              />
              <Input
                allowClear
                placeholder={t('exchangeRates.source')}
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
              {t('exchangeRates.sync')}
            </Button>
          )}
        />
      )}
    >
      <ContentCard>
        {loading && rates.length === 0 ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : filteredRates.length === 0 ? (
          <Empty description={hasFilters ? t('common.noMatch') : t('common.noData')} />
        ) : (
          <Table dataSource={filteredRates} columns={columns} rowKey="id" size="small" pagination={{ pageSize: 50 }} />
        )}
      </ContentCard>
    </PageLayout>
  );
};

export default ExchangeRatesPage;
