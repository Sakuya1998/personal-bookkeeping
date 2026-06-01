import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import type { Ledger } from '../../api/types';

const mockGet = vi.fn();

vi.mock('../../api/client', () => ({
  default: {
    get: mockGet,
  },
}));

const mockLedger: Ledger = {
  id: 'ledger-1',
  user_id: 'user-1',
  name: '测试账本',
  description: null,
  base_currency: 'CNY',
  icon: '📒',
  color: null,
  is_archived: false,
  sort_order: 0,
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
};

const mockSetCurrentLedger = vi.fn();

vi.mock('../../store/appStore', () => ({
  useAppStore: vi.fn(() => ({
    currentLedger: mockLedger,
    ledgers: [mockLedger],
    setCurrentLedger: mockSetCurrentLedger,
  })),
}));

let CalendarViewPage: React.FC;

describe('CalendarViewPage', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const mod = await import('../../pages/CalendarViewPage');
    CalendarViewPage = mod.default;
  });

  it('renders with PageLayout and no in-page ledger select', async () => {
    let resolveGet: ((value: unknown) => void) | null = null;
    mockGet.mockImplementation(() => new Promise((r) => {
      resolveGet = r;
    }));

    const { container } = render(
      <MemoryRouter initialEntries={['/ledgers/ledger-1/calendar']}>
        <Routes>
          <Route path="/ledgers/:ledger_id/calendar" element={<CalendarViewPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText('日历视图')).toBeInTheDocument();
    expect(container.querySelector('.ant-select')).toBeNull();

    await act(async () => {
      resolveGet?.({ data: { code: 200, data: [], message: 'ok' } });
    });

    await act(async () => {});
  });
});
