export interface User {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
}

export interface Ledger {
  id: string;
  user_id: string;
  name: string;
  description: string | null;
  base_currency: string;
  icon: string | null;
  color: string | null;
  is_archived: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  user_id: string;
  ledger_id: string | null;
  name: string;
  type: 'income' | 'expense';
  icon: string | null;
  color: string | null;
  parent_id: string | null;
  sort_order: number;
  is_active: boolean;
  children?: Category[];
}

export interface Transaction {
  id: string;
  ledger_id: string;
  user_id: string;
  category_id: string;
  type: 'income' | 'expense';
  amount: number;
  currency: string;
  exchange_rate: number;
  base_amount: number;
  description: string | null;
  transaction_date: string;
  tags: string | null;
  is_reconciled: boolean;
  created_at: string;
  updated_at: string;
  category?: Category;
}

export interface ExchangeRate {
  id: string;
  from_currency: string;
  to_currency: string;
  rate: number;
  date: string;
  source: string | null;
  created_at: string;
}

export interface ApiResponse<T> {
  code: number;
  data: T;
  message: string;
}

export interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface LedgerSummary {
  total_income: number;
  total_expense: number;
  balance: number;
  base_currency: string;
  expense_by_category: {
    category_id: string;
    category_name: string;
    category_icon: string;
    total: number;
    count: number;
  }[];
}

export interface AuthResponse {
  token: string;
  user: User;
}
