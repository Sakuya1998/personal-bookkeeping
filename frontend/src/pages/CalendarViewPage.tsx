import React, { useEffect, useState, useMemo } from 'react';
import { Card, Button, Row, Col, Select, Spin, Empty, Tag } from 'antd';
import { LeftOutlined, RightOutlined } from '@ant-design/icons';
import { useParams } from 'react-router-dom';
import client from '../api/client';
import { ApiResponse, DailyTransactionItem, Transaction } from '../api/types';
import { useAppStore } from '../store/appStore';
import { formatCurrency } from '../utils/currency';
import dayjs, { Dayjs } from 'dayjs';

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日'];

const CalendarViewPage: React.FC = () => {
  const { currentLedger, ledgers, setCurrentLedger } = useAppStore();
  const { ledger_id } = useParams<{ ledger_id: string }>();

  const [currentMonth, setCurrentMonth] = useState<Dayjs>(dayjs().startOf('month'));
  const [dailyData, setDailyData] = useState<DailyTransactionItem[]>([]);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [dayTxns, setDayTxns] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);

  const ledgerId = currentLedger?.id;

  // Sync ledger from URL param
  useEffect(() => {
    if (ledger_id && ledgers.length > 0 && currentLedger?.id !== ledger_id) {
      const ledger = ledgers.find((l) => l.id === ledger_id);
      if (ledger) setCurrentLedger(ledger);
    }
  }, [ledger_id, ledgers, currentLedger, setCurrentLedger]);

  // Fetch daily data — defer all setState to microtasks
  useEffect(() => {
    if (!ledgerId) return;
    queueMicrotask(() => {
      setLoading(true);
      setDailyData([]);
      setSelectedDate(null);
      setDayTxns([]);
    });
    client
      .get<ApiResponse<DailyTransactionItem[]>>(
        `/ledgers/${ledgerId}/daily-transactions?year=${currentMonth.year()}&month=${currentMonth.month() + 1}`,
      )
      .then((res) => setDailyData(res.data.data || []))
      .finally(() => setLoading(false));
  }, [ledgerId, currentMonth]);

  // Fetch transactions for selected date
  const handleDateClick = (dateStr: string) => {
    if (!ledgerId || !dateStr) return;
    setSelectedDate(dateStr);
    client
      .get<ApiResponse<{ items: Transaction[] }>>(
        `/ledgers/${ledgerId}/transactions?start_date=${dateStr}&end_date=${dateStr}&page_size=50`,
      )
      .then((res) => setDayTxns(res.data.data?.items || []));
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

  if (!currentLedger) {
    return <Empty description="请先选择账本" />;
  }

  return (
    <div>
      {/* Ledger selector + month navigation */}
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col>
          <Select
            value={currentLedger.id}
            onChange={(id) => {
              const ledger = ledgers.find((l) => l.id === id);
              if (ledger) setCurrentLedger(ledger);
            }}
            style={{ width: 200 }}
            options={ledgers.map((l) => ({ label: `${l.icon || ''} ${l.name}`, value: l.id }))}
          />
        </Col>
        <Col>
          <Button icon={<LeftOutlined />} onClick={prevMonth} />
          <span style={{ margin: '0 16px', fontSize: 16, fontWeight: 600, verticalAlign: 'middle' }}>
            {currentMonth.format('YYYY 年 M 月')}
          </span>
          <Button icon={<RightOutlined />} onClick={nextMonth} />
        </Col>
      </Row>

      <Spin spinning={loading}>
        {/* Calendar grid */}
        <Card>
          {/* Weekday headers */}
          <Row style={{ borderBottom: '2px solid #f0f0f0', paddingBottom: 8, marginBottom: 8 }}>
            {WEEKDAYS.map((wd) => (
              <Col span={3} key={wd} style={{ textAlign: 'center', fontWeight: 600, padding: '4px 0' }}>
                <span style={{ color: wd === '六' || wd === '日' ? '#ff4d4f' : undefined }}>{wd}</span>
              </Col>
            ))}
          </Row>

          {/* Calendar rows */}
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
                        <div style={{ color: '#52c41a' }}>+{formatCurrency(cell.item.income, currentLedger.base_currency)}</div>
                      )}
                      {cell.item.expense > 0 && (
                        <div style={{ color: '#ff4d4f' }}>-{formatCurrency(cell.item.expense, currentLedger.base_currency)}</div>
                      )}
                      {cell.item.count > 1 && (
                        <div style={{ fontSize: 11, color: '#bbb' }}>{cell.item.count} 笔</div>
                      )}
                    </div>
                  )}
                </Col>
              ))}
            </Row>
          ))}
        </Card>

        {/* Selected day details */}
        {selectedDate && (
          <Card
            title={`${selectedDate} 交易详情${selectedItem ? ` — 收入 ${formatCurrency(selectedItem.income, currentLedger.base_currency)} / 支出 ${formatCurrency(selectedItem.expense, currentLedger.base_currency)}` : ''}`}
            style={{ marginTop: 16 }}
          >
            {dayTxns.length === 0 ? (
              <Empty description="当日无交易" image={Empty.PRESENTED_IMAGE_SIMPLE} />
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
                      {txn.type === 'income' ? '收入' : '支出'}
                    </Tag>
                    <span>{txn.category?.icon || ''} {txn.category?.name || '未知分类'}</span>
                    {txn.description && <span style={{ marginLeft: 8, color: '#999' }}>{txn.description}</span>}
                  </div>
                  <span
                    style={{
                      fontWeight: 600,
                      color: txn.type === 'income' ? '#52c41a' : '#ff4d4f',
                    }}
                  >
                    {txn.type === 'income' ? '+' : '-'}
                    {formatCurrency(txn.base_amount, currentLedger.base_currency)}
                  </span>
                </div>
              ))
            )}
          </Card>
        )}
      </Spin>
    </div>
  );
};

export default CalendarViewPage;
