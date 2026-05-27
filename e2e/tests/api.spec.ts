import { test, expect, APIRequestContext } from '@playwright/test';
import {
  registerUser,
  loginUser,
  authContext,
  expectOK,
  testHelpers,
  TEST_USER,
} from './helpers.js';

let api: APIRequestContext;
let token: string;
let ledgerId: string;
let categoryId: string;
let transactionId: string;

test.describe('API E2E', () => {
  // ─────────────────────── Health ───────────────────────

  test('GET /api/v1/health returns ok', async ({ request }) => {
    const res = await request.get('/api/v1/health');
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.status).toBe('ok');
    expect(body.db).toBeDefined();
  });

  // ─────────────────────── Auth ────────────────────────

  test.describe('Auth', () => {
    test('POST /api/v1/auth/register creates user', async ({ request }) => {
      const username = `e2e_${Date.now()}`;
      const res = await request.post('/api/v1/auth/register', {
        data: {
          username,
          email: `${username}@test.com`,
          password: 'test123456',
        },
      });
      expect(res.status()).toBe(201);
      const body = await res.json();
      expectOK(res.status(), body);
      expect(body.data.token).toBeTruthy();
      expect(body.data.user.username).toBe(username);
    });

    test('POST /api/v1/auth/login returns token', async ({ request }) => {
      const res = await request.post('/api/v1/auth/login', {
        data: {
          username: TEST_USER.username,
          password: TEST_USER.password,
        },
      });

      if (res.status() === 401) {
        // 用户不存在，先注册
        const reg = await request.post('/api/v1/auth/register', {
          data: TEST_USER,
        });
        const regBody = await reg.json();
        expect(reg.status()).toBe(201);
        token = regBody.data.token;
      } else {
        expect(res.status()).toBe(200);
        const body = await res.json();
        token = body.data.token;
      }
    });

    test('GET /api/v1/auth/me returns current user', async () => {
      const { token: t } = await loginUser(
        test.info().project.use.baseURL!,
      );
      api = await authContext(test.info().project.use.baseURL!, t);

      const res = await api.get('/api/v1/auth/me');
      const body = await res.json();
      expectOK(res.status(), body);
      expect(body.data.username).toBe(TEST_USER.username);
    });

    test('PUT /api/v1/auth/password changes password', async () => {
      const res = await api.put('/api/v1/auth/password', {
        data: { old_password: TEST_USER.password, new_password: 'newpass789' },
      });
      expectOK(res.status(), await res.json());

      // 恢复密码
      await api.put('/api/v1/auth/password', {
        data: { old_password: 'newpass789', new_password: TEST_USER.password },
      });
    });
  });

  // ─────────────────────── Ledgers ──────────────────────

  test.describe('Ledgers', () => {
    test('POST /api/v1/ledgers creates ledger', async () => {
      const { id, name } = await testHelpers.createLedger(api, 'E2E-Main');
      ledgerId = id;
      expect(name).toContain('E2E-Main');
    });

    test('GET /api/v1/ledgers lists ledgers', async () => {
      const res = await api.get('/api/v1/ledgers');
      const body = await res.json();
      expectOK(res.status(), body);
      expect(Array.isArray(body.data)).toBe(true);
      expect(body.data.some((l: any) => l.id === ledgerId)).toBe(true);
    });

    test('GET /api/v1/ledgers/:id returns single ledger', async () => {
      const res = await api.get(`/api/v1/ledgers/${ledgerId}`);
      const body = await res.json();
      expectOK(res.status(), body);
      expect(body.data.id).toBe(ledgerId);
    });

    test('PUT /api/v1/ledgers/:id updates ledger', async () => {
      const res = await api.put(`/api/v1/ledgers/${ledgerId}`, {
        data: { name: 'E2E-Main-Renamed' },
      });
      expectOK(res.status(), await res.json());

      // 恢复
      await api.put(`/api/v1/ledgers/${ledgerId}`, {
        data: { name: 'E2E-Main' },
      });
    });

    test('GET /api/v1/ledgers/:id/summary returns summary', async () => {
      const res = await api.get(`/api/v1/ledgers/${ledgerId}/summary`);
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.data).toBeDefined();
    });
  });

  // ─────────────────────── Categories ───────────────────

  test.describe('Categories', () => {
    test('POST /api/v1/categories creates category', async () => {
      const { id, name } = await testHelpers.createCategory(
        api,
        ledgerId,
        'E2E-Food',
      );
      categoryId = id;
      expect(name).toBe('E2E-Food');
    });

    test('GET /api/v1/ledgers/:id/categories lists categories', async () => {
      const res = await api.get(
        `/api/v1/ledgers/${ledgerId}/categories`,
      );
      const body = await res.json();
      expectOK(res.status(), body);
      expect(Array.isArray(body.data)).toBe(true);
    });

    test('PUT /api/v1/categories/:id updates category', async () => {
      const res = await api.put(`/api/v1/categories/${categoryId}`, {
        data: { name: 'E2E-Food-Updated' },
      });
      expectOK(res.status(), await res.json());
    });
  });

  // ─────────────────────── Transactions ─────────────────

  test.describe('Transactions', () => {
    test('POST /api/v1/transactions creates transaction', async () => {
      const { id } = await testHelpers.createTransaction(
        api,
        ledgerId,
        categoryId,
        {
          description: 'E2E lunch',
          amount: 35.5,
        },
      );
      transactionId = id;
    });

    test('GET /api/v1/ledgers/:id/transactions lists transactions', async () => {
      const res = await api.get(
        `/api/v1/ledgers/${ledgerId}/transactions?page=1&page_size=10`,
      );
      const body = await res.json();
      expectOK(res.status(), body);
      const list = body.data?.list || body.data || [];
      expect(Array.isArray(list)).toBe(true);
      expect(list.some((t: any) => t.id === transactionId)).toBe(true);
    });

    test('PUT /api/v1/transactions/:id updates transaction', async () => {
      const res = await api.put(`/api/v1/transactions/${transactionId}`, {
        data: { description: 'E2E lunch updated', amount: 42.0 },
      });
      expectOK(res.status(), await res.json());
    });

    test('GET /api/v1/ledgers/:id/report generates report', async () => {
      const res = await api.get(
        `/api/v1/ledgers/${ledgerId}/report?start_date=2024-01-01&end_date=2030-12-31`,
      );
      const body = await res.json();
      expectOK(res.status(), body);
    });

    test('DELETE /api/v1/transactions/:id removes transaction', async () => {
      const res = await api.delete(
        `/api/v1/transactions/${transactionId}`,
      );
      expectOK(res.status(), await res.json());
    });
  });

  // ─────────────────────── Cleanup ──────────────────────

  test.afterAll(async () => {
    if (ledgerId) {
      await api.delete(`/api/v1/ledgers/${ledgerId}`);
    }
  });
});
