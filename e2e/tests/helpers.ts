import { APIRequestContext, expect, request } from '@playwright/test';

export const TEST_USER = {
  username: process.env.TEST_USERNAME || 'e2e_test',
  email: process.env.TEST_EMAIL || 'e2e_test@example.com',
  password: process.env.TEST_PASSWORD || 'test123456',
};

export interface AuthSession {
  token: string;
  userId: string;
}

/** 注册并返回 token + user id */
export async function registerUser(baseURL: string): Promise<AuthSession> {
  const ctx = await request.newContext({ baseURL });
  const res = await ctx.post('/api/v1/auth/register', {
    data: TEST_USER,
  });
  const body = await res.json();

  if (res.status() === 409) {
    // 用户已存在，改为登录
    return loginUser(baseURL);
  }

  expect(res.status()).toBe(201);
  expect(body.data).toBeDefined();
  expect(body.data.token).toBeTruthy();
  return {
    token: body.data.token,
    userId: body.data.user.id,
  };
}

/** 登录并返回 token + user id */
export async function loginUser(baseURL: string): Promise<AuthSession> {
  const ctx = await request.newContext({ baseURL });
  const res = await ctx.post('/api/v1/auth/login', {
    data: {
      username: TEST_USER.username,
      password: TEST_USER.password,
    },
  });
  const body = await res.json();
  expect(res.status()).toBe(200);
  expect(body.data.token).toBeTruthy();
  return {
    token: body.data.token,
    userId: body.data.user.id,
  };
}

/** 创建认证上下文（注入 Authorization header） */
export async function authContext(
  baseURL: string,
  token: string,
): Promise<APIRequestContext> {
  return request.newContext({
    baseURL,
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  });
}

/** 解析统一响应体 `{ code, message, data }` */
export function expectOK(status: number, body: any) {
  expect(status).toBeGreaterThanOrEqual(200);
  expect(status).toBeLessThan(300);
  expect(body.code).toBe(status);
  expect(body.message).toBe('ok');
}

async function createLedger(
  ctx: APIRequestContext,
  name?: string,
): Promise<{ id: string; name: string }> {
  const ledgerName = name || `E2E-Ledger-${Date.now()}`;
  const res = await ctx.post('/api/v1/ledgers', {
    data: { name: ledgerName, base_currency: 'CNY' },
  });
  const body = await res.json();
  expectOK(res.status(), body);
  return { id: body.data.id, name: ledgerName };
}

async function createCategory(
  ctx: APIRequestContext,
  ledgerId: string,
  name?: string,
): Promise<{ id: string; name: string }> {
  const catName = name || `E2E-Cat-${Date.now()}`;
  const res = await ctx.post('/api/v1/categories', {
    data: {
      ledger_id: ledgerId,
      name: catName,
      type: 'expense',
    },
  });
  const body = await res.json();
  expectOK(res.status(), body);
  return { id: body.data.id, name: catName };
}

async function createTransaction(
  ctx: APIRequestContext,
  ledgerId: string,
  categoryId: string,
  overrides?: Partial<{
    type: string;
    amount: number;
    description: string;
    currency: string;
  }>,
): Promise<{ id: string }> {
  const res = await ctx.post('/api/v1/transactions', {
    data: {
      ledger_id: ledgerId,
      category_id: categoryId,
      type: overrides?.type || 'expense',
      amount: overrides?.amount || 29.9,
      description: overrides?.description || 'E2E test transaction',
      currency: overrides?.currency || 'CNY',
      transaction_date: new Date().toISOString().slice(0, 10),
    },
  });
  const body = await res.json();
  expectOK(res.status(), body);
  return { id: body.data.transaction.id };
}

export const testHelpers = {
  createLedger,
  createCategory,
  createTransaction,
};
