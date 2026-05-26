/* eslint-disable @typescript-eslint/no-explicit-any */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
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

const mockCurrentLedger = {
  id: 'ledger-1',
  name: '测试账本',
  base_currency: 'CNY',
  description: null,
  icon: null,
  color: null,
  is_archived: false,
  sort_order: 0,
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
  user_id: 'user-1',
};

vi.mock('../../store/appStore', () => ({
  useAppStore: vi.fn(() => ({
    currentLedger: mockCurrentLedger,
  })),
}));

// Note: we intentionally do NOT mock dayjs — antd's DatePicker depends on
// the real dayjs for internal date parsing/validation (.isValid(), etc.).

// ---------------------------------------------------------------------------
// Helpers: factory functions for API responses
// ---------------------------------------------------------------------------

function makeApiResponse<T>(data: T) {
  return { data: { code: 200, data, message: 'ok' } };
}

function makeBudget(overrides?: Partial<Record<string, any>>) {
  return {
    id: 'budget-1',
    user_id: 'user-1',
    ledger_id: 'ledger-1',
    category_id: null,
    month: '2025-06',
    amount: 5000,
    created_at: '2025-06-01T00:00:00Z',
    updated_at: '2025-06-01T00:00:00Z',
    ...overrides,
  };
}

function makeCategory(overrides?: Partial<Record<string, any>>) {
  return {
    id: 'cat-1',
    user_id: 'user-1',
    ledger_id: 'ledger-1',
    name: '餐饮',
    type: 'expense',
    icon: '🍔',
    color: null,
    parent_id: null,
    sort_order: 1,
    is_active: true,
    ...overrides,
  };
}

function makeStatusItem(overrides?: Partial<Record<string, any>>) {
  return {
    budget_id: 'budget-1',
    category_id: null,
    name: '全部支出',
    icon: null,
    budget: 5000,
    spent: 2000,
    percentage: 40,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// The component under test
// ---------------------------------------------------------------------------

let BudgetPage: React.FC;

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('BudgetPage', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const mod = await import('../../pages/BudgetPage');
    BudgetPage = mod.default;
  });

  it('shows loading skeleton while fetching data', async () => {
    mockGet.mockReturnValue(new Promise(() => {}));

    render(<BudgetPage />);

    await waitFor(() => {
      expect(screen.getByText('预算执行状态')).toBeInTheDocument();
    });
  });

  it('shows empty states when no budgets exist', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      if (url.includes('/budgets/status')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });

    render(<BudgetPage />);

    await waitFor(() => {
      expect(screen.getByText('本月暂无预算')).toBeInTheDocument();
    });

    expect(screen.getByText('暂无预算设置')).toBeInTheDocument();
  });

  it('renders budget status items and budget table when data is available', async () => {
    const budgets = [
      makeBudget({ id: 'b1', category_id: 'cat-1', amount: 3000 }),
      makeBudget({ id: 'b2', category_id: null, amount: 5000 }),
    ];
    const statusItems = [
      makeStatusItem({ budget_id: 'b1', name: '餐饮', budget: 3000, spent: 1500, percentage: 50 }),
      makeStatusItem({ budget_id: 'b2', name: '全部支出', budget: 5000, spent: 5000, percentage: 100 }),
    ];
    const categories = [makeCategory({ id: 'cat-1', name: '餐饮', icon: '🍔' })];

    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse(categories));
      }
      if (url.includes('/budgets/status')) {
        return Promise.resolve(makeApiResponse(statusItems));
      }
      return Promise.resolve(makeApiResponse(budgets));
    });

    render(<BudgetPage />);

    // Wait for status items to appear
    await waitFor(() => {
      expect(screen.getByText('餐饮')).toBeInTheDocument();
    });

    // "全部支出" appears both in the status section and as a Tag for budget with null category
    const allExpenseElements = screen.getAllByText('全部支出');
    expect(allExpenseElements.length).toBeGreaterThanOrEqual(2);

    // Budget table — category name from joined category
    expect(screen.getByText('🍔 餐饮')).toBeInTheDocument();

    // Amount formatting (formatCurrency uses toFixed, no comma separators)
    expect(screen.getByText('¥5000.00')).toBeInTheDocument();
    expect(screen.getByText('¥3000.00')).toBeInTheDocument();

    // Budget list table header
    expect(screen.getByText('预算设置')).toBeInTheDocument();
  });

  it('opens the create modal when the "新增预算" button is clicked', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      if (url.includes('/budgets/status')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });

    render(<BudgetPage />);

    await waitFor(() => {
      expect(screen.getByText('本月暂无预算')).toBeInTheDocument();
    });

    const createBtn = screen.getByRole('button', { name: /新增预算/ });
    fireEvent.click(createBtn);

    await waitFor(() => {
      expect(screen.getByText('预算金额')).toBeInTheDocument();
    });
  });

  it('calls client.post with correct data when the form is submitted', async () => {
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      if (url.includes('/budgets/status')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse([]));
    });
    mockPost.mockResolvedValue(makeApiResponse({}));

    render(<BudgetPage />);

    await waitFor(() => {
      expect(screen.getByText('本月暂无预算')).toBeInTheDocument();
    });

    // Open modal
    fireEvent.click(screen.getByRole('button', { name: /新增预算/ }));

    await waitFor(() => {
      expect(screen.getByText('预算金额')).toBeInTheDocument();
    });

    // Fill in the amount field
    const amountInput = screen.getByPlaceholderText('例如 5000');
    fireEvent.change(amountInput, { target: { value: '3000' } });

    // Submit form via its onFinish handler — click the Modal OK button
    const okButton = screen.getByRole('button', { name: 'OK' });
    fireEvent.click(okButton);

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith('/budgets', {
        ledger_id: 'ledger-1',
        category_id: null,
        month: expect.any(String),
        amount: 3000,
      });
    });
  });

  it('calls client.delete when the delete button is clicked and confirmed', async () => {
    const budgets = [makeBudget({ id: 'b1', amount: 2000 })];
    mockGet.mockImplementation((url: string) => {
      if (url.includes('/categories')) {
        return Promise.resolve(makeApiResponse([]));
      }
      if (url.includes('/budgets/status')) {
        return Promise.resolve(makeApiResponse([]));
      }
      return Promise.resolve(makeApiResponse(budgets));
    });
    mockDelete.mockResolvedValue(makeApiResponse({}));

    render(<BudgetPage />);

    await waitFor(() => {
      expect(screen.getByText('¥2000.00')).toBeInTheDocument();
    });

    // Click delete button
    const deleteBtn = screen.getByRole('button', { name: /delete/i });
    fireEvent.click(deleteBtn);

    // Wait for Popconfirm to render the "确定" button
    await waitFor(() => {
      expect(screen.getByText('确定删除？')).toBeInTheDocument();
    });

    // Click the confirm button in Popconfirm — antd uses "OK" in English locale (jsdom)
    const confirmBtn = screen.getByRole('button', { name: 'OK' });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(mockDelete).toHaveBeenCalledWith('/budgets/b1');
    });
  });
});
