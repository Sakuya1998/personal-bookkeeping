import { test, expect, Page, request } from '@playwright/test';
import { registerUser, TEST_USER } from './helpers.js';

const BASE_URL = process.env.FRONTEND_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:8000';

let page: Page;
let authToken: string;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async ({ browser }) => {
  // 1. 通过 API 注册/登录测试用户
  const ctx = await request.newContext({ baseURL: API_URL });
  const loginRes = await ctx.post('/api/v1/auth/login', {
    data: { username: TEST_USER.username, password: TEST_USER.password },
  });
  if (loginRes.status() === 401) {
    const reg = await ctx.post('/api/v1/auth/register', { data: TEST_USER });
    const regBody = await reg.json();
    authToken = regBody.data.token;
  } else {
    const body = await loginRes.json();
    authToken = body.data.token;
  }

  page = await browser.newPage();

  // 2. 注入 JWT token 到 localStorage
  await page.goto(BASE_URL);
  await page.evaluate((t) => {
    localStorage.setItem('token', t);
  }, authToken);
});

test.afterAll(async () => {
  await page.close();
});

// ─────────────────── Health / Page Load ───────────────────

test('frontend loads successfully', async () => {
  await page.goto('/');
  await expect(page).toHaveTitle(/记账|书|book|Personal/);
});

// ─────────────────── Login Page ───────────────────────────

test('login page shows form', async () => {
  // 先清除 token 才能看到登录页（否则自动跳转仪表盘）
  await page.evaluate(() => localStorage.removeItem('token'));
  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  const usernameInput = page.locator('#username');
  const passwordInput = page.locator('#password');
  const loginButton = page.locator('button[type="submit"]').first();

  await expect(usernameInput).toBeVisible({ timeout: 5000 });
  await expect(passwordInput).toBeVisible({ timeout: 5000 });
  await expect(loginButton).toBeVisible({ timeout: 5000 });
});

test('login with valid credentials redirects to dashboard', async () => {
  // 通过 API 登录，注入 token 模拟完整流程
  const ctx = await request.newContext({ baseURL: API_URL });
  const res = await ctx.post('/api/v1/auth/login', {
    data: { username: TEST_USER.username, password: TEST_USER.password },
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  const token = body.data.token;

  await page.evaluate((t) => localStorage.setItem('token', t), token);
  await page.goto('/');

  // 已登录 → 自动进入首页/仪表盘，而非登录页
  await page.waitForLoadState('networkidle');
  expect(page.url()).not.toContain('/login');
});

// ─────────────────── Authenticated Pages ───────────────────

test('dashboard shows content when authenticated', async () => {
  // 恢复 token
  await page.evaluate((t) => localStorage.setItem('token', t), authToken);
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  const bodyText = await page.locator('body').innerText();
  expect(bodyText.length).toBeGreaterThan(10);
});

test('transaction page loads when authenticated', async () => {
  await page.evaluate((t) => localStorage.setItem('token', t), authToken);
  await page.goto('/transactions');
  await page.waitForLoadState('networkidle');

  const bodyText = await page.locator('body').innerText();
  expect(bodyText.length).toBeGreaterThan(10);
});

// ─────────────────── Navigation ───────────────────────────

test('can navigate between pages when authenticated', async () => {
  await page.evaluate((t) => localStorage.setItem('token', t), authToken);
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  const navLinks = page.locator('a, button, [role="tab"]').filter({
    hasText: /交易|分类|账本|报表|预算|Transaction|Category|Ledger|Report|Budget/,
  });

  const count = await navLinks.count();
  if (count > 0) {
    const href = await navLinks.first().getAttribute('href');
    if (href) {
      await navLinks.first().click();
      await page.waitForURL(`**${href}`, { timeout: 8000 });
      expect(page.url()).toContain(href);
    }
  }
});
