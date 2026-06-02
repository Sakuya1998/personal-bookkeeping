import { test, expect, APIRequestContext } from '@playwright/test';
import {
  registerUser,
  loginUser,
  authContext,
  expectOK,
  testHelpers,
  TEST_USER,
} from './helpers.js';

interface Ctx {
  api: APIRequestContext;
  ledgerId: string;
  categoryId: string;
  transactionId: string;
  budgetId: string;
  memberApi: APIRequestContext;
  memberUserId: string;
  recurringId: string;
  exchangeRateId: string;
}

const ctx = {} as Ctx;

// 使用 serial 确保测试顺序执行
test.describe.configure({ mode: 'serial' });

test.beforeAll(async ({ request }) => {
  // 1. 确保测试用户存在
  const loginRes = await request.post('/api/v1/auth/login', {
    data: { username: TEST_USER.username, password: TEST_USER.password },
  });
  let token: string;
  if (loginRes.status() === 401) {
    const reg = await request.post('/api/v1/auth/register', { data: TEST_USER });
    const regBody = await reg.json();
    expect(reg.status()).toBe(201);
    token = regBody.data.token;
  } else {
    const body = await loginRes.json();
    token = body.data.token;
  }

  ctx.api = await authContext(test.info().project.use.baseURL!, token);

  // 2. 创建账本
  const { id: lid } = await testHelpers.createLedger(ctx.api, 'E2E-Test');
  ctx.ledgerId = lid;

  // 3. 创建分类
  const { id: cid } = await testHelpers.createCategory(ctx.api, lid, 'E2E-Test-Cat');
  ctx.categoryId = cid;

  // 4. 创建交易
  const { id: tid } = await testHelpers.createTransaction(ctx.api, lid, cid, {
    description: 'E2E-init',
    amount: 10,
  });
  ctx.transactionId = tid;
});

// ─────────────────────── Health ───────────────────────

test.describe('Health', () => {
  test('GET /api/v1/health returns ok', async ({ request }) => {
    const res = await request.get('/api/v1/health');
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.status).toBe('ok');
    expect(body.db).toBeDefined();
  });
});

// ─────────────────────── Auth ────────────────────────

test.describe('Auth', () => {
  test('POST /api/v1/auth/register creates fresh user', async ({ request }) => {
    const username = `e2e_${Date.now()}`;
    const res = await request.post('/api/v1/auth/register', {
      data: { username, email: `${username}@test.com`, password: 'test123456' },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.token).toBeTruthy();
    expect(body.data.user.username).toBe(username);
  });

  test('GET /api/v1/auth/me returns current user', async () => {
    const res = await ctx.api.get('/api/v1/auth/me');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(body.data.username).toBe(TEST_USER.username);
  });

  test('PUT /api/v1/auth/password round-trips', async () => {
    // change to new
    let res = await ctx.api.put('/api/v1/auth/password', {
      data: { old_password: TEST_USER.password, new_password: 'newpass789' },
    });
    expectOK(res.status(), await res.json());

    // restore
    res = await ctx.api.put('/api/v1/auth/password', {
      data: { old_password: 'newpass789', new_password: TEST_USER.password },
    });
    expectOK(res.status(), await res.json());
  });
});

// ─────────────────────── Ledgers ──────────────────────

test.describe('Ledgers', () => {
  test('GET /api/v1/ledgers lists ledgers', async () => {
    const res = await ctx.api.get('/api/v1/ledgers');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.data.some((l: any) => l.id === ctx.ledgerId)).toBe(true);
  });

  test('GET /api/v1/ledgers/:id returns ledger', async () => {
    const res = await ctx.api.get(`/api/v1/ledgers/${ctx.ledgerId}`);
    const body = await res.json();
    expectOK(res.status(), body);
    expect(body.data.id).toBe(ctx.ledgerId);
  });

  test('PUT /api/v1/ledgers/:id updates ledger', async () => {
    let res = await ctx.api.put(`/api/v1/ledgers/${ctx.ledgerId}`, {
      data: { name: 'E2E-Renamed' },
    });
    expectOK(res.status(), await res.json());

    res = await ctx.api.put(`/api/v1/ledgers/${ctx.ledgerId}`, {
      data: { name: 'E2E-Test' },
    });
    expectOK(res.status(), await res.json());
  });

  test('GET /api/v1/ledgers/:id/summary returns summary', async () => {
    const res = await ctx.api.get(`/api/v1/ledgers/${ctx.ledgerId}/summary`);
    const body = await res.json();
    expectOK(res.status(), body);
    expect(body.data).toBeDefined();
  });
});

// ─────────────────────── Categories ───────────────────

test.describe('Categories', () => {
  test('POST /api/v1/categories creates category', async () => {
    const { id } = await testHelpers.createCategory(ctx.api, ctx.ledgerId, 'E2E-Cat-2');
    expect(id).toBeTruthy();
  });

  test('GET /api/v1/ledgers/:id/categories lists categories', async () => {
    const res = await ctx.api.get(`/api/v1/ledgers/${ctx.ledgerId}/categories`);
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
  });

  test('PUT /api/v1/categories/:id updates category', async () => {
    const res = await ctx.api.put(`/api/v1/categories/${ctx.categoryId}`, {
      data: { name: 'E2E-Cat-Updated' },
    });
    expectOK(res.status(), await res.json());
  });
});

// ─────────────────────── Transactions ─────────────────

test.describe('Transactions', () => {
  test('POST /api/v1/transactions creates transaction', async () => {
    const { id } = await testHelpers.createTransaction(ctx.api, ctx.ledgerId, ctx.categoryId, {
      description: 'E2E lunch',
      amount: 35.5,
    });
    ctx.transactionId = id;
  });

  test('GET /api/v1/ledgers/:id/transactions lists transactions', async () => {
    const res = await ctx.api.get(
      `/api/v1/ledgers/${ctx.ledgerId}/transactions?page=1&page_size=10`,
    );
    const body = await res.json();
    expectOK(res.status(), body);
    // data 是 { items, total, page, page_size, total_pages }
    expect(Array.isArray(body.data.items)).toBe(true);
    expect(body.data.items.some((t: any) => t.id === ctx.transactionId)).toBe(true);
  });

  test('PUT /api/v1/transactions/:id updates transaction', async () => {
    const res = await ctx.api.put(`/api/v1/transactions/${ctx.transactionId}`, {
      data: { description: 'E2E lunch updated', amount: 42.0 },
    });
    expectOK(res.status(), await res.json());
  });

  test('GET /api/v1/ledgers/:id/report/preview returns JSON', async () => {
    const res = await ctx.api.get(
      `/api/v1/ledgers/${ctx.ledgerId}/report/preview?period=monthly&date=2024-01`,
    );
    const body = await res.json();
    expectOK(res.status(), body);
  });

  test('DELETE /api/v1/transactions/:id removes transaction', async () => {
    const res = await ctx.api.delete(`/api/v1/transactions/${ctx.transactionId}`);
    expectOK(res.status(), await res.json());
  });
});

// ─────────────────────── Budgets ────────────────────────

test.describe('Budgets', () => {
  test('POST /api/v1/budgets creates budget', async () => {
    const res = await ctx.api.post('/api/v1/budgets', {
      data: {
        ledger_id: ctx.ledgerId,
        category_id: ctx.categoryId,
        month: '2026-06',
        amount: 2000,
      },
    });
    const body = await res.json();
    expectOK(res.status(), body);
    ctx.budgetId = body.data.id;
  });

  test('GET /api/v1/budgets lists budgets', async () => {
    const res = await ctx.api.get('/api/v1/budgets');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.data.some((b: any) => b.id === ctx.budgetId)).toBe(true);
  });

  test('GET /api/v1/budgets/status returns budget status', async () => {
    const res = await ctx.api.get('/api/v1/budgets/status');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
  });

  test('DELETE /api/v1/budgets/:id removes budget', async () => {
    const res = await ctx.api.delete(`/api/v1/budgets/${ctx.budgetId}`);
    const body = await res.json();
    expectOK(res.status(), body);
  });
});

// ─────────────────────── Members ────────────────────────

test.describe('Members', () => {
  test('invite member by username', async () => {
    // Register a second user via API directly
    const memberUsername = `mem_${Date.now()}`;
    const regRes = await ctx.api.post('/api/v1/auth/register', {
      data: {
        username: memberUsername,
        email: `${memberUsername}@test.com`,
        password: 'test123456',
      },
    });
    const regBody = await regRes.json();
    expectOK(regRes.status(), regBody);
    ctx.memberUserId = regBody.data.user.id;

    // Create authenticated context for the member
    ctx.memberApi = await authContext(
      test.info().project.use.baseURL!,
      regBody.data.token,
    );

    // Owner invites member
    const res = await ctx.api.post(`/api/v1/ledgers/${ctx.ledgerId}/members`, {
      data: { username: memberUsername },
    });
    const body = await res.json();
    expectOK(res.status(), body);
  });

  test('GET /api/v1/ledgers/:id/members lists members', async () => {
    const res = await ctx.api.get(`/api/v1/ledgers/${ctx.ledgerId}/members`);
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.data.some((m: any) => m.user_id === ctx.memberUserId)).toBe(true);
  });

  test('member can access shared ledger', async () => {
    const res = await ctx.memberApi.get(`/api/v1/ledgers/${ctx.ledgerId}`);
    const body = await res.json();
    expectOK(res.status(), body);
    expect(body.data.id).toBe(ctx.ledgerId);
  });

  test('member cannot create category with member role', async () => {
    const res = await ctx.memberApi.post('/api/v1/categories', {
      data: {
        ledger_id: ctx.ledgerId,
        name: 'Should-Fail',
        type: 'expense',
      },
    });
    // Member role → expect 403 Forbidden
    expect(res.status()).toBe(403);
  });

  test('remove member', async () => {
    const res = await ctx.api.delete(
      `/api/v1/ledgers/${ctx.ledgerId}/members/${ctx.memberUserId}`,
    );
    const body = await res.json();
    expectOK(res.status(), body);

    // Verify member can no longer access
    const check = await ctx.memberApi.get(`/api/v1/ledgers/${ctx.ledgerId}`);
    expect(check.status()).toBe(404);
  });
});

// ─────────────────────── Recurring ──────────────────────

test.describe('Recurring', () => {
  test('POST /api/v1/recurring creates rule', async () => {
    const res = await ctx.api.post('/api/v1/recurring', {
      data: {
        ledger_id: ctx.ledgerId,
        category_id: ctx.categoryId,
        type: 'expense',
        amount: 500,
        currency: 'CNY',
        frequency: 'monthly',
        start_date: '2026-06-01',
      },
    });
    const body = await res.json();
    expectOK(res.status(), body);
    ctx.recurringId = body.data.id;
  });

  test('GET /api/v1/recurring lists rules', async () => {
    const res = await ctx.api.get('/api/v1/recurring');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
    expect(body.data.some((r: any) => r.id === ctx.recurringId)).toBe(true);
  });

  test('PUT /api/v1/recurring/:id updates rule', async () => {
    const res = await ctx.api.put(`/api/v1/recurring/${ctx.recurringId}`, {
      data: { amount: 800 },
    });
    expectOK(res.status(), await res.json());
  });

  test('GET /api/v1/recurring/upcoming returns upcoming', async () => {
    const res = await ctx.api.get('/api/v1/recurring/upcoming');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
  });

  test('DELETE /api/v1/recurring/:id removes rule', async () => {
    const res = await ctx.api.delete(`/api/v1/recurring/${ctx.recurringId}`);
    expectOK(res.status(), await res.json());
  });
});

// ─────────────────────── Exchange Rates ─────────────────

test.describe('Exchange Rates', () => {
  test('GET /api/v1/exchange-rates lists rates', async () => {
    const res = await ctx.api.get('/api/v1/exchange-rates');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
  });

  test('GET /api/v1/exchange-rates/latest returns latest', async () => {
    const res = await ctx.api.get('/api/v1/exchange-rates/latest');
    const body = await res.json();
    expectOK(res.status(), body);
    expect(Array.isArray(body.data)).toBe(true);
  });

  test('POST /api/v1/exchange-rates/sync can be called', async () => {
    // Sync may fail if no API key is configured, but should not 5xx
    const res = await ctx.api.post('/api/v1/exchange-rates/sync');
    if (res.status() !== 200) {
      const body = await res.json();
      expect(body.code).toBeGreaterThanOrEqual(400);
      expect(body.code).toBeLessThan(500);
      expect(body.message).toBeTruthy();
    }
  });
});

test.afterAll(async () => {
  if (ctx.ledgerId) {
    await ctx.api.delete(`/api/v1/ledgers/${ctx.ledgerId}`);
  }
});
