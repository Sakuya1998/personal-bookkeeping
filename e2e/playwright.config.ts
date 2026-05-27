import { defineConfig } from '@playwright/test';

const API_URL = process.env.API_URL || 'http://localhost:8000';
const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:3000';
const TEST_USERNAME = process.env.TEST_USERNAME || 'e2e_test';
const TEST_PASSWORD = process.env.TEST_PASSWORD || 'test123456';
const TEST_EMAIL = process.env.TEST_EMAIL || 'e2e_test@example.com';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  expect: {
    timeout: 10000,
  },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : 1,
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'playwright-report/results.json' }],
  ],
  use: {
    baseURL: FRONTEND_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
    },
  },
  projects: [
    {
      name: 'api',
      testMatch: '**/api.spec.ts',
      use: {
        baseURL: API_URL,
      },
    },
    {
      name: 'browser',
      testMatch: '**/browser.spec.ts',
      use: {
        baseURL: FRONTEND_URL,
        browserName: 'chromium',
        headless: !process.env.HEADED,
      },
    },
  ],
});
