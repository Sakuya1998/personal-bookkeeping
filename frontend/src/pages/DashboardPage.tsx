import React, { useEffect, useState } from 'react';
import {
  Row, Col, Card, Statistic, Empty, Button, Spin, Modal, Select,
  DatePicker, message, Space, Skeleton,
} from 'antd';
import {
  ArrowUpOutlined, ArrowDownOutlined, WalletOutlined, DownloadOutlined, FileTextOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import ReactECharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import dayjs from 'dayjs';
import client from '../api/client';
import { ApiResponse, LedgerSummary, MonthlyTrendItem, CategoryBreakdownItem } from '../api/types';
import { useAppStore } from '../store/appStore';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';

const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const { currentLedger } = useAppStore();
  const [summary, setSummary] = useState<LedgerSummary | null>(null);
  const [monthlyTrend, setMonthlyTrend] = useState<MonthlyTrendItem[]>([]);
  const [categoryBreakdown, setCategoryBreakdown] = useState<CategoryBreakdownItem[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  // Export modal state
  const [exportOpen, setExportOpen] = useState(false);
  const [exportFormat, setExportFormat] = useState<'csv' | 'json'>('csv');
  const [exportDateRange, setExportDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  // Report modal state
  const [reportOpen, setReportOpen] = useState(false);
  const [reportMonth, setReportMonth] = useState<dayjs.Dayjs>(dayjs());
  const [reportPeriod, setReportPeriod] = useState<'monthly' | 'quarterly' | 'yearly'>('monthly');
  const [reportLoading, setReportLoading] = useState(false);

  useEffect(() => {
    if (!currentLedger) return;
    queueMicrotask(() => setLoading(true));
    Promise.all([
      client.get<ApiResponse<LedgerSummary>>(`/ledgers/${currentLedger.id}/summary`),
      client.get<ApiResponse<MonthlyTrendItem[]>>(`/ledgers/${currentLedger.id}/monthly-trend?months=6`),
      client.get<ApiResponse<CategoryBreakdownItem[]>>(`/ledgers/${currentLedger.id}/category-breakdown`),
    ])
      .then(([sumRes, trendRes, catRes]) => {
        setSummary(sumRes.data.data);
        setMonthlyTrend(trendRes.data.data || []);
        setCategoryBreakdown(catRes.data.data || []);
      })
      .catch(err => console.error(t('dashboard.fetchDataFailed'), err))
      .finally(() => setLoading(false));
  }, [currentLedger, t]);

  const handleExport = async () => {
    if (!currentLedger) return;
    const params = new URLSearchParams({ format: exportFormat });
    if (exportDateRange) {
      params.set('start_date', exportDateRange[0].format('YYYY-MM-DD'));
      params.set('end_date', exportDateRange[1].format('YYYY-MM-DD'));
    }
    try {
      const url = `/ledgers/${currentLedger.id}/export?${params}`;
      const res = await client.get(url, { responseType: 'text' });
      const contentType = res.headers['content-type'] as string || '';

      if (contentType.includes('text/csv')) {
        // File download — create blob and trigger download
        const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' });
        const downloadUrl = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = downloadUrl;
        a.download = `export.${exportFormat}`;
        a.click();
        URL.revokeObjectURL(downloadUrl);
        message.success(t('common.exportSuccess'));
      } else if (contentType.includes('application/json')) {
        // Async task response or JSON data
        try {
          const parsed = JSON.parse(res.data);
          if (parsed.data?.task_id) {
            message.info(t('dashboard.exportTaskSubmitted', { taskId: parsed.data.task_id, total: parsed.data.total }));
          } else if (Array.isArray(parsed.data)) {
            // Direct JSON export — download
            const blob = new Blob([JSON.stringify(parsed.data, null, 2)], { type: 'application/json' });
            const downloadUrl = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = `export.json`;
            a.click();
            URL.revokeObjectURL(downloadUrl);
            message.success(t('common.exportSuccess'));
          } else {
            message.info(JSON.stringify(parsed));
          }
        } catch {
          // plain text response, just download
          const blob = new Blob([res.data], { type: 'text/plain' });
          const downloadUrl = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = downloadUrl;
          a.download = `export.txt`;
          a.click();
          URL.revokeObjectURL(downloadUrl);
        }
      }
      setExportOpen(false);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('common.exportFailed'));
    }
  };

  const handleGenerateReport = async () => {
    if (!currentLedger) return;
    setReportLoading(true);
    try {
      const date = reportPeriod === 'yearly' ? reportMonth.format('YYYY') : reportMonth.format('YYYY-MM');
      const res = await client.get(`/ledgers/${currentLedger.id}/report`, {
        params: { period: reportPeriod, date },
        responseType: 'blob',
      });
      const blob = new Blob([res.data], { type: 'application/pdf' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${currentLedger.name}_${date}_report.pdf`;
      a.click();
      URL.revokeObjectURL(url);
      message.success(t('dashboard.reportSuccess'));
      setReportOpen(false);
    } catch (err: unknown) {
      const apiErr = err as { response?: { data?: { message?: string } } };
      message.error(apiErr.response?.data?.message || t('dashboard.reportFailed'));
    } finally {
      setReportLoading(false);
    }
  };

  if (!currentLedger) {
    return (
      <PageLayout header={<PageTitle title={t('dashboard.title')} />}>
        <Empty description={t('dashboard.noLedger')}>
          <Button type="primary" onClick={() => navigate('/ledgers')}>{t('dashboard.createLedger')}</Button>
        </Empty>
      </PageLayout>
    );
  }

  const trendOption: EChartsOption = {
    tooltip: { trigger: 'axis' },
    legend: { data: [t('transactions.income'), t('transactions.expense')], top: 0 },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: monthlyTrend.map((m) => m.month) },
    yAxis: { type: 'value' },
    series: [
      {
        name: t('transactions.income'), type: 'line', smooth: true,
        data: monthlyTrend.map((m) => m.income),
        itemStyle: { color: '#52c41a' },
        areaStyle: { color: 'rgba(82,196,26,0.1)' },
      },
      {
        name: t('transactions.expense'), type: 'line', smooth: true,
        data: monthlyTrend.map((m) => m.expense),
        itemStyle: { color: '#ff4d4f' },
        areaStyle: { color: 'rgba(255,77,79,0.08)' },
      },
    ],
  };

  const expenseItems = categoryBreakdown.filter((c) => c.type === 'expense');

  const ringOption: EChartsOption = {
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', right: 0, top: 'middle', data: expenseItems.map((c) => c.category_name) },
    series: [{
      name: t('dashboard.expenseByCategory'), type: 'pie', radius: ['45%', '72%'], center: ['40%', '50%'],
      avoidLabelOverlap: false, padAngle: 2,
      itemStyle: { borderRadius: 6 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: expenseItems.map((c) => ({ value: c.total, name: c.category_name })),
    }],
  };

  return (
    <PageLayout
      header={<PageTitle title={t('dashboard.title')} description={t('dashboard.currentLedger', { name: currentLedger.name })} />}
      toolbar={(
        <PageToolbar
          right={(
            <>
              <Button type="primary" icon={<FileTextOutlined />} onClick={() => setReportOpen(true)}>
                {t('dashboard.generateReport')}
              </Button>
              <Button icon={<DownloadOutlined />} onClick={() => setExportOpen(true)}>
                {t('common.export')}
              </Button>
            </>
          )}
        />
      )}
    >
      <Spin spinning={loading}>
        {loading && !summary ? (
          <>
            <Row gutter={16} style={{ marginBottom: 24 }}>
              {[1, 2, 3].map((i) => (
                <Col span={8} key={i}>
                  <Card><Skeleton active paragraph={{ rows: 1 }} title={{ width: '60%' }} /></Card>
                </Col>
              ))}
            </Row>
            <Row gutter={16}>
              <Col xs={24} lg={12} style={{ marginBottom: 16 }}>
                <Card title={t('dashboard.monthlyTrend')}><Skeleton active paragraph={{ rows: 6 }} /></Card>
              </Col>
              <Col xs={24} lg={12} style={{ marginBottom: 16 }}>
                <Card title={t('dashboard.expenseByCategory')}><Skeleton active paragraph={{ rows: 6 }} /></Card>
              </Col>
            </Row>
          </>
        ) : (
          <>
            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <Card hoverable onClick={() => navigate('/transactions?type=income')}>
                  <Statistic
                    title={t('dashboard.totalIncome')} value={summary?.total_income || 0} precision={2}
                    prefix={<ArrowUpOutlined />} valueStyle={{ color: '#52c41a' }}
                    suffix={currentLedger.base_currency}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card hoverable onClick={() => navigate('/transactions?type=expense')}>
                  <Statistic
                    title={t('dashboard.totalExpense')} value={summary?.total_expense || 0} precision={2}
                    prefix={<ArrowDownOutlined />} valueStyle={{ color: '#ff4d4f' }}
                    suffix={currentLedger.base_currency}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    title={t('dashboard.balance')} value={summary?.balance || 0} precision={2}
                    prefix={<WalletOutlined />}
                    valueStyle={{ color: (summary?.balance || 0) >= 0 ? '#1890ff' : '#ff4d4f' }}
                    suffix={currentLedger.base_currency}
                  />
                </Card>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col xs={24} lg={12} style={{ marginBottom: 16 }}>
                <Card title={t('dashboard.monthlyTrend')}>
                  {monthlyTrend.length > 0 ? (
                    <ReactECharts option={trendOption} style={{ height: 320 }} />
                  ) : (
                    <Empty description={t('common.noData')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
                  )}
                </Card>
              </Col>
              <Col xs={24} lg={12} style={{ marginBottom: 16 }}>
                <Card title={t('dashboard.expenseByCategory')}>
                  {expenseItems.length > 0 ? (
                    <ReactECharts option={ringOption} style={{ height: 320 }} />
                  ) : (
                    <Empty description={t('common.noData')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
                  )}
                </Card>
              </Col>
            </Row>
          </>
        )}

        <Modal
          title={t('dashboard.exportData')}
          open={exportOpen}
          onOk={handleExport}
          onCancel={() => setExportOpen(false)}
          okText={t('dashboard.startExport')}
          cancelText={t('common.cancel')}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <div style={{ marginBottom: 8 }}>{t('dashboard.exportFormat')}</div>
              <Select
                value={exportFormat}
                onChange={(v) => setExportFormat(v)}
                style={{ width: '100%' }}
                options={[
                  { label: t('dashboard.csvFormat'), value: 'csv' },
                  { label: 'JSON', value: 'json' },
                ]}
              />
            </div>
            <div>
              <div style={{ marginBottom: 8 }}>{t('dashboard.exportDateRangeHint')}</div>
              <DatePicker.RangePicker
                style={{ width: '100%' }}
                value={exportDateRange}
                onChange={(dates) => setExportDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
              />
            </div>
          </Space>
        </Modal>

        <Modal
          title={t('dashboard.reportTitle')}
          open={reportOpen}
          onOk={handleGenerateReport}
          onCancel={() => setReportOpen(false)}
          confirmLoading={reportLoading}
          okText={t('dashboard.downloadPdf')}
          cancelText={t('common.cancel')}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <div style={{ marginBottom: 8 }}>{t('dashboard.reportPeriod')}</div>
              <Select
                value={reportPeriod}
                onChange={(v) => setReportPeriod(v)}
                style={{ width: '100%' }}
                options={[
                  { label: t('dashboard.monthly'), value: 'monthly' },
                  { label: t('dashboard.quarterly'), value: 'quarterly' },
                  { label: t('dashboard.yearly'), value: 'yearly' },
                ]}
              />
            </div>
            <div>
              <div style={{ marginBottom: 8 }}>
                {reportPeriod === 'monthly' ? t('dashboard.selectMonth') : reportPeriod === 'yearly' ? t('dashboard.selectYear') : t('dashboard.selectQuarter')}
              </div>
              <DatePicker
                picker={reportPeriod === 'yearly' ? 'year' : 'month'}
                value={reportMonth}
                onChange={(d) => d && setReportMonth(d)}
                style={{ width: '100%' }}
              />
            </div>
          </Space>
        </Modal>
      </Spin>
    </PageLayout>
  );
};

export default DashboardPage;
