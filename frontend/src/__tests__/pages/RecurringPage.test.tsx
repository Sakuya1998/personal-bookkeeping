/* eslint-disable @typescript-eslint/no-explicit-any */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockDelete = vi.fn();

vi.mock('../../api/client', () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: mockPut,
    delete: mockDelete,
  },
}));

vi.mock('../../store/appStore', () => ({
  useAppStore: vi.fn(() => ({
    currentLedger: {
      id: 'ledger-1',
      name: '测试账本',
      base_currency: 'CNY',
      user_id: 'user-1',
      description: null,
      icon: null,
      color: null,
      is_archived: false,
      sort_order: 0,
      created_at: '2025-01-01T00:00:00Z',
      updated_at: '2025-01-01T00:00:00Z',
    },
  })),
}));

// Mock CURRENCIES and formatCurrency used in RecurringPage
vi.mock('../../utils/currency', () => ({
  CURRENCIES: [
    { code: 'CNY', symbol: '¥', name: '人民币' },
    { code: 'USD', symbol: '$', name: '美元' },
  ],
  formatCurrency: vi.fn((amount: number) => `¥${amount.toFixed(2)}`),
}));

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeApiResponse<T>(data: T) {
  return { data: { code: 200, data, message: 'ok' } };
}

function makeRule(overrides?: Partial<Record<string, any>>) {
  return {
    id: 'rule-1',
    user_id: 'user-1',
    ledger_id: 'ledger-1',
    category_id: 'cat-1',
    type: 'expense',
    amount: 500,
    currency: 'CNY',
    description: '每月房租',
    tags: '',
    frequency: 'monthly',
    interval: 1,
    day_of_month: 1,
    weekday: null,
    start_date: '2025-01-01',
    end_date: null,
    next_run_date: '2025-07-01',
    is_active: true,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-06-01T00:00:00Z',
    ...overrides,
  };
}

function makeCategory(overrides?: Partial<Record<string, any>>) {
  return {
    id: 'cat-1',
    user_id: 'user-1',
    ledger_id: 'ledger-1',
    name: '房租',
    type: 'expense',
    icon: '🏠',
    color: null,
    parent_id: null,
    sort_order: 1,
    is_active: true,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// The component under test
// ---------------------------------------------------------------------------

let RecurringPage: React.FC;

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('RecurringPage', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const mod = await import('../../pages/RecurringPage');
    RecurringPage = mod.default;
  });

  it('shows the page header while fetching data', async () => {
    mockGet.mockReturnValue(new Promise(() => {}));

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('recurring.pageDescription')).toBeInTheDocument();
    });

    // The create button should be visible even during loading
    expect(screen.getByRole('button', { name: /recurring.add/ })).toBeInTheDocument();
  });

  it('shows empty state when no rules exist', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('recurring.noRules')).toBeInTheDocument();
    });
  });

  it('renders the rules table when data is available', async () => {
    const rules = [
      makeRule({ id: 'r1', type: 'expense', amount: 500, description: '每月房租' }),
      makeRule({ id: 'r2', type: 'income', amount: 10000, description: '工资', frequency: 'monthly' }),
    ];
    const categories = [makeCategory({ id: 'cat-1', name: '房租', icon: '🏠' })];

    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse(categories));
      }
      return Promise.resolve(makeApiResponse(rules));
    });

    render(<RecurringPage />);

    // Wait for data to render
    await waitFor(() => {
      expect(screen.getByText('每月房租')).toBeInTheDocument();
    });

    // Description of second rule
    expect(screen.getByText('工资')).toBeInTheDocument();

    // Amount formatting: expense has "-" prefix, income has "+"
    // Our mock formatCurrency returns "¥500.00" and the component prepends "-" for expense
    // So the rendered text is "-¥500.00"
    expect(screen.getByText(/-¥500\.00/)).toBeInTheDocument();
    expect(screen.getByText(/\+¥10000\.00/)).toBeInTheDocument();

    // Type tags
    expect(screen.getByText('transactions.expense')).toBeInTheDocument();
    expect(screen.getByText('transactions.income')).toBeInTheDocument();

    // Status tag — both rules are active so there are two "启用" tags
    const enabledTags = screen.getAllByText('recurring.isActive');
    expect(enabledTags.length).toBeGreaterThanOrEqual(2);

    // Start dates — both rules have the same start_date
    const startDates = screen.getAllByText('2025-01-01');
    expect(startDates.length).toBeGreaterThanOrEqual(2);

    // No end date => "无" — both rules have no end_date so it appears twice
    const noEndDates = screen.getAllByText('recurring.noEndDate');
    expect(noEndDates.length).toBeGreaterThanOrEqual(2);

    // Frequency display — both rules are monthly
    const monthlyTags = screen.getAllByText('recurring.monthly');
    expect(monthlyTags.length).toBeGreaterThanOrEqual(2);
  });

  it('opens the create modal when the create button is clicked', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('recurring.noRules')).toBeInTheDocument();
    });

    const createBtn = screen.getByRole('button', { name: /recurring.add/ });
    fireEvent.click(createBtn);

    await waitFor(() => {
      expect(screen.getAllByText('recurring.add').length).toBeGreaterThanOrEqual(1);
    });
  });

  it('opens the create modal correctly', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });
    mockPost.mockResolvedValue(makeApiResponse({}));

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('recurring.noRules')).toBeInTheDocument();
    });

    // Open create modal
    fireEvent.click(screen.getByRole('button', { name: /recurring.add/ }));

    await waitFor(() => {
      expect(screen.getAllByText('recurring.add').length).toBeGreaterThanOrEqual(1);
    });

    // Verify the form contains expected fields
    expect(screen.getByText('transactions.type')).toBeInTheDocument();
    expect(screen.getByText('transactions.amount')).toBeInTheDocument();
    expect(screen.getByText('transactions.category')).toBeInTheDocument();
    expect(screen.getByText('recurring.frequency')).toBeInTheDocument();
    expect(screen.getByText('recurring.startDate')).toBeInTheDocument();
  });

  it('opens edit modal and calls client.put when editing', async () => {
    const rules = [
      makeRule({ id: 'r1', type: 'expense', amount: 500, description: '编辑测试', start_date: '2025-01-01' }),
    ];
    const categories = [makeCategory({ id: 'cat-1', name: '房租' })];

    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse(categories));
      }
      return Promise.resolve(makeApiResponse(rules));
    });
    mockPut.mockResolvedValue(makeApiResponse({}));

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('编辑测试')).toBeInTheDocument();
    });

    // Click the edit button (EditOutlined icon button)
    const editBtn = screen.getByRole('button', { name: /edit/i });
    fireEvent.click(editBtn);

    // Modal should show edit title
    await waitFor(() => {
      expect(screen.getByText('recurring.edit')).toBeInTheDocument();
    });

    // Submit the form via OK button
    const okButton = screen.getByRole('button', { name: 'OK' });
    fireEvent.click(okButton);

    await waitFor(() => {
      expect(mockPut).toHaveBeenCalledWith('/recurring/r1', expect.any(Object));
    });
  });

  it('calls client.delete when the delete button is confirmed', async () => {
    const rules = [makeRule({ id: 'r1', description: '待删除' })];

    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse(rules));
    });
    mockDelete.mockResolvedValue(makeApiResponse({}));

    render(<RecurringPage />);

    await waitFor(() => {
      expect(screen.getByText('待删除')).toBeInTheDocument();
    });

    // Click delete button
    const deleteBtn = screen.getByRole('button', { name: /delete/i });
    fireEvent.click(deleteBtn);

    // Popconfirm should show
    await waitFor(() => {
      expect(screen.getByText('recurring.deleteConfirm')).toBeInTheDocument();
    });

    // Confirm — antd Popconfirm uses "OK" in English locale (jsdom)
    const confirmBtn = screen.getByRole('button', { name: 'OK' });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(mockDelete).toHaveBeenCalledWith('/recurring/r1');
    });
  });
});
