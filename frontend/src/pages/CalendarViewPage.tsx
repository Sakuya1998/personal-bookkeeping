import React, { useEffect, useState, useMemo, useRef } from 'react';
import { Card, Button, Row, Col, Spin, Empty, Tag, message } from 'antd';
import { LeftOutlined, RightOutlined } from '@ant-design/icons';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import client from '../api/client';
import { ApiResponse, DailyTransactionItem, Transaction } from '../api/types';
import { useAppStore } from '../store/appStore';
import { formatCurrency } from '../utils/currency';
import dayjs, { Dayjs } from 'dayjs';
import PageLayout from '../components/layout/PageLayout';
import PageTitle from '../components/layout/PageTitle';
import PageToolbar from '../components/layout/PageToolbar';
import ContentCard from '../components/layout/ContentCard';

const CalendarViewPage: React.FC = () => {
  const { t } = useTranslation();
  const { currentLedger, ledgers, setCurrentLedger } = useAppStore();
  const { ledger_id } = useParams<{ ledger_id: string }>();

  const [currentMonth, setCurrentMonth] = useState<Dayjs>(dayjs().startOf('month'));
  const [dailyData, setDailyData] = useState<DailyTransactionItem[]>([]);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [dayTxns, setDayTxns] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);

  const dateClickFetchRef = useRef<{ cancelled: boolean }>({ cancelled: false });

  const urlLedgerId = ledger_id || '';

  const ledgerFromUrl = useMemo(() => {
    if (!urlLedgerId) return null;
    return ledgers.find((l) => l.id === urlLedgerId) || null;
  }, [ledgers, urlLedgerId]);

  useEffect(() => {
    if (!ledgerFromUrl) return;
    if (currentLedger?.id === urlLedgerId) return;
    setCurrentLedger(ledgerFromUrl);
  }, [currentLedger?.id, ledgerFromUrl, setCurrentLedger, urlLedgerId]);

  // Fetch daily data — defer all setState to microtasks
  useEffect(() => {
    if (!urlLedgerId) return;
    let cancelled = false;
    queueMicrotask(() => {
      if (cancelled) return;
      setLoading(true);
      setDailyData([]);
      setSelectedDate(null);
      setDayTxns([]);
    });
    client
      .get<ApiResponse<DailyTransactionItem[]>>(
        `/ledgers/${urlLedgerId}/daily-transactions?year=${currentMonth.year()}&month=${currentMonth.month() + 1}`,
      )
      .then((res) => { if (!cancelled) setDailyData(res.data.data || []); })
      .catch(err => { if (cancelled) return; console.error(t('calendar.fetchDailyDataFailed'), err); message.error(t('calendar.fetchDailyDataFailed')); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [urlLedgerId, currentMonth]);

  // Fetch transactions for selected date
  const handleDateClick = (dateStr: string) => {
    if (!urlLedgerId || !dateStr) return;
    dateClickFetchRef.current.cancelled = true;
    const current = { cancelled: false };
    dateClickFetchRef.current = current;
    setSelectedDate(dateStr);
    client
      .get<ApiResponse<{ items: Transaction[] }>>(
        `/ledgers/${urlLedgerId}/transactions?start_date=${dateStr}&end_date=${dateStr}&page_size=50`,
      )
      .then((res) => { if (!current.cancelled) setDayTxns(res.data.data?.items || []); })
      .catch(err => { if (current.cancelled) return; console.error(t('calendar.fetchDateTransactionsFailed'), err); message.error(t('calendar.fetchDateTransactionsFailed')); });
  };

  // Build calendar data map
  const dateMap = useMemo(() => {
    const map = new Map<string, DailyTransactionItem>();
    for (const d of dailyData) {
      map.set(d.date, d);
    }
    return map;
  }, [dailyData]);

  // Calendar grid data
  const calendarCells = useMemo(() => {
    const startOfMonth = currentMonth;
    const daysInMonth = startOfMonth.daysInMonth();
    const startDayOfWeek = startOfMonth.day();
    // Convert Sunday(0) to Monday(1) based: Mon=0, Sun=6
    const startOffset = startDayOfWeek === 0 ? 6 : startDayOfWeek - 1;

    const cells: { dateStr: string; day: number; item: DailyTransactionItem | undefined; isCurrentMonth: boolean }[] = [];

    // Previous month padding
    const prevMonthEnd = startOfMonth.subtract(1, 'day').date();
    for (let i = startOffset - 1; i >= 0; i--) {
      const date = startOfMonth.subtract(i + 1, 'day');
      cells.push({
        dateStr: date.format('YYYY-MM-DD'),
        day: prevMonthEnd - i,
        item: undefined,
        isCurrentMonth: false,
      });
    }

    // Current month
    for (let d = 1; d <= daysInMonth; d++) {
      const date = startOfMonth.date(d);
      const dateStr = date.format('YYYY-MM-DD');
      cells.push({
        dateStr,
        day: d,
        item: dateMap.get(dateStr),
        isCurrentMonth: true,
      });
    }

    // Next month padding to fill 42 cells (6 rows)
    const remaining = 42 - cells.length;
    for (let d = 1; d <= remaining; d++) {
      const date = startOfMonth.add(1, 'month').date(d);
      cells.push({
        dateStr: date.format('YYYY-MM-DD'),
        day: d,
        item: undefined,
        isCurrentMonth: false,
      });
    }

    return cells;
  }, [currentMonth, dateMap]);

  const prevMonth = () => setCurrentMonth(currentMonth.subtract(1, 'month'));
  const nextMonth = () => setCurrentMonth(currentMonth.add(1, 'month'));

  const selectedItem = selectedDate ? dateMap.get(selectedDate) : undefined;

  const weekdays = [t('calendar.mon'), t('calendar.tue'), t('calendar.wed'), t('calendar.thu'), t('calendar.fri'), t('calendar.sat'), t('calendar.sun')];

  if (!urlLedgerId || !ledgerFromUrl) {
    return (
      <PageLayout header={<PageTitle title={t('calendar.title')} />}>
        <Empty description={t('calendar.noLedger')} />
      </PageLayout>
    );
  }

  return (
    <PageLayout
      header={<PageTitle title={t('calendar.title')} description={t('dashboard.currentLedger', { name: ledgerFromUrl.name })} />}
      toolbar={(
        <PageToolbar
          left={<span style={{ fontSize: 16, fontWeight: 600 }}>{currentMonth.format(t('calendar.monthFormat'))}</span>}
          right={(
            <>
              <Button icon={<LeftOutlined />} onClick={prevMonth} />
              <Button icon={<RightOutlined />} onClick={nextMonth} />
            </>
          )}
        />
      )}
    >
      <Spin spinning={loading}>
        <ContentCard>
          <Row style={{ borderBottom: '2px solid #f0f0f0', paddingBottom: 8, marginBottom: 8 }}>
            {weekdays.map((wd, idx) => (
              <Col span={3} key={wd} style={{ textAlign: 'center', fontWeight: 600, padding: '4px 0' }}>
                <span style={{ color: idx >= 5 ? '#ff4d4f' : undefined }}>{wd}</span>
              </Col>
            ))}
          </Row>

          {[0, 1, 2, 3, 4, 5].map((week) => (
            <Row key={week} style={{ minHeight: 90 }}>
              {calendarCells.slice(week * 7, week * 7 + 7).map((cell) => (
                <Col
                  span={3}
                  key={cell.dateStr}
                  style={{
                    border: '1px solid #f0f0f0',
                    padding: '4px 6px',
                    minHeight: 88,
                    cursor: cell.isCurrentMonth ? 'pointer' : 'default',
                    background: cell.dateStr === selectedDate ? '#e6f7ff' : undefined,
                    opacity: cell.isCurrentMonth ? 1 : 0.35,
                  }}
                  onClick={() => cell.isCurrentMonth && handleDateClick(cell.dateStr)}
                >
                  <div style={{ textAlign: 'right', fontSize: 12, color: '#999', marginBottom: 4 }}>
                    {cell.day}
                  </div>
                  {cell.item && (
                    <div style={{ fontSize: 13, lineHeight: '18px' }}>
                      {cell.item.income > 0 && (
                        <div style={{ color: '#52c41a' }}>+{formatCurrency(cell.item.income, ledgerFromUrl.base_currency)}</div>
                      )}
                      {cell.item.expense > 0 && (
                        <div style={{ color: '#ff4d4f' }}>-{formatCurrency(cell.item.expense, ledgerFromUrl.base_currency)}</div>
                      )}
                      {cell.item.count > 1 && (
                        <div style={{ fontSize: 11, color: '#bbb' }}>{t('calendar.transactionCount', { count: cell.item.count })}</div>
                      )}
                    </div>
                  )}
                </Col>
              ))}
            </Row>
          ))}
        </ContentCard>

        {selectedDate && (
          <Card
            title={`${t('calendar.transactionDetails', { date: selectedDate })}${selectedItem ? ` — ${t('calendar.incomeExpense', { income: formatCurrency(selectedItem.income, ledgerFromUrl.base_currency), expense: formatCurrency(selectedItem.expense, ledgerFromUrl.base_currency) })}` : ''}`}
            style={{ marginTop: 16 }}
          >
            {dayTxns.length === 0 ? (
              <Empty description={t('calendar.noTransactions')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              dayTxns.map((txn) => (
                <div
                  key={txn.id}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '8px 0',
                    borderBottom: '1px solid #f5f5f5',
                  }}
                >
                  <div>
                    <Tag color={txn.type === 'income' ? 'green' : 'red'}>
                      {txn.type === 'income' ? t('transactions.income') : t('transactions.expense')}
                    </Tag>
                    <span>{txn.category?.icon || ''} {txn.category?.name || t('categories.unknown')}</span>
                    {txn.description && <span style={{ marginLeft: 8, color: '#999' }}>{txn.description}</span>}
                  </div>
                  <span
                    style={{
                      fontWeight: 600,
                      color: txn.type === 'income' ? '#52c41a' : '#ff4d4f',
                    }}
                  >
                    {txn.type === 'income' ? '+' : '-'}
                    {formatCurrency(txn.base_amount, ledgerFromUrl.base_currency)}
                  </span>
                </div>
              ))
            )}
          </Card>
        )}
      </Spin>
    </PageLayout>
  );
};

export default CalendarViewPage;
