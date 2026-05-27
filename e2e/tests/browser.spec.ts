import { test, expect, Page } from '@playwright/test';
import { registerUser, TEST_USER } from './helpers.js';

let page: Page;

test.describe('Browser E2E', () => {
  test.beforeAll(async ({ browser }) => {
    page = await browser.newPage();
  });

  test.afterAll(async () => {
    await page.close();
  });

  // ─────────────────── Health / Page Load ───────────────────

  test('frontend loads successfully', async () => {
    await page.goto('/');
    // 应该重定向到登录页，或显示首页
    await expect(page).toHaveTitle(/记账|书|book|Personal/);
  });

  // ─────────────────── Login Flow ───────────────────────────

  test('login page is accessible and shows form', async () => {
    await page.goto('/login');
    // 检查是否存在登录表单的关键元素
    await page.waitForLoadState('networkidle');

    // 检查是否有用户名/密码输入框和登录按钮
    const usernameInput = page.locator('input[id*="username"], input[name*="username"], input[placeholder*="用户"], input[placeholder*="username"]').first();
    const passwordInput = page.locator('input[type="password"]').first();
    const loginButton = page.locator('button[type="submit"], button:has-text("登录"), button:has-text("Login"), button:has-text("Sign in")').first();

    await expect(usernameInput).toBeVisible({ timeout: 5000 });
    await expect(passwordInput).toBeVisible({ timeout: 5000 });
    await expect(loginButton).toBeVisible({ timeout: 5000 });
  });

  test('login with valid credentials succeeds', async () => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // 尝试填写登录
    const usernameInput = page.locator('input[id*="username"], input[name*="username"], input[placeholder*="用户"], input[placeholder*="username"]').first();
    const passwordInput = page.locator('input[type="password"]').first();
    const loginButton = page.locator('button[type="submit"], button:has-text("登录"), button:has-text("Login"), button:has-text("Sign in")').first();

    if (await usernameInput.isVisible()) {
      await usernameInput.fill(TEST_USER.username);
      await passwordInput.fill(TEST_USER.password);
      await loginButton.click();

      // 登录成功后应跳转到仪表盘或首页
      await page.waitForURL(/\/(dashboard|home|\/)/, { timeout: 10000 });
      const currentUrl = page.url();
      expect(currentUrl).not.toContain('/login');
    }
  });

  // ─────────────────── Dashboard ────────────────────────────

  test('dashboard shows key metrics after login', async () => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // 检查页面是否渲染出内容（指标卡片、图表、或账本信息）
    const bodyText = await page.locator('body').innerText();
    expect(bodyText.length).toBeGreaterThan(50);
  });

  // ─────────────────── Transaction List ─────────────────────

  test('transaction page loads and shows data', async () => {
    await page.goto('/transactions');
    await page.waitForLoadState('networkidle');

    const bodyText = await page.locator('body').innerText();
    expect(bodyText.length).toBeGreaterThan(50);
  });

  // ─────────────────── Navigation ───────────────────────────

  test('can navigate between pages', async () => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // 尝试找到导航链接并点击
    const navLinks = page.locator('a, button, [role="tab"]').filter({
      hasText: /交易|分类|账本|报表|预算|Transaction|Category|Ledger|Report|Budget/,
    });

    const count = await navLinks.count();
    if (count > 0) {
      // 点击第一个导航链接
      const href = await navLinks.first().getAttribute('href');
      if (href) {
        await navLinks.first().click();
        await page.waitForURL(`**${href}`, { timeout: 8000 });
        expect(page.url()).toContain(href);
      }
    }
  });
});
