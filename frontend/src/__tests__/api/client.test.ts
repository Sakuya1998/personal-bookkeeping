import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

// Track interceptor handlers for later testing
let requestHandler: ((config: any) => any) | null = null;
let responseSuccessHandler: ((res: any) => any) | null = null;
let responseErrorHandler: ((err: any) => any) | null = null;

// Mock axios before importing client
vi.mock('axios', () => {
  const mockAxiosInstance = {
    defaults: {},
    interceptors: {
      request: {
        use: vi.fn((handler: any) => {
          requestHandler = handler;
          return 0;
        }),
        eject: vi.fn(),
        clear: vi.fn(),
      },
      response: {
        use: vi.fn((onFulfilled: any, onRejected: any) => {
          responseSuccessHandler = onFulfilled;
          responseErrorHandler = onRejected;
          return 0;
        }),
        eject: vi.fn(),
        clear: vi.fn(),
      },
    },
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    request: vi.fn(),
  };

  return {
    default: {
      create: vi.fn(() => mockAxiosInstance),
    },
  };
});

describe('API client', () => {
  beforeEach(() => {
    localStorage.clear();
    requestHandler = null;
    responseSuccessHandler = null;
    responseErrorHandler = null;
  });

  afterEach(() => {
    vi.resetModules();
  });

  describe('request interceptor', () => {
    it('injects Authorization header when token exists in localStorage', async () => {
      localStorage.setItem('token', 'test-token-123');

      // Import triggers interceptor registration
      await import('../../api/client');

      const config = { headers: {} };
      const result = requestHandler!(config);

      expect(result.headers.Authorization).toBe('Bearer test-token-123');
    });

    it('does not add Authorization header when no token in localStorage', async () => {
      // localStorage already cleared in beforeEach
      await import('../../api/client');

      const config = { headers: {} };
      const result = requestHandler!(config);

      expect(result.headers.Authorization).toBeUndefined();
    });

    it('returns the config object unchanged (aside from headers)', async () => {
      localStorage.setItem('token', 'my-token');
      await import('../../api/client');

      const config = { url: '/test', method: 'get', headers: {} };
      const result = requestHandler!(config);

      expect(result.url).toBe('/test');
      expect(result.method).toBe('get');
    });
  });

  describe('response success handler', () => {
    it('passes through the response on success', async () => {
      await import('../../api/client');

      const response = { data: { code: 200, data: {} }, status: 200 };
      const result = responseSuccessHandler!(response);

      expect(result).toBe(response);
    });
  });

  describe('response error handler', () => {
    it('removes token and redirects to /login on 401', async () => {
      localStorage.setItem('token', 'should-be-cleared');
      await import('../../api/client');

      // Mock window.location.href
      const originalLocation = window.location;
      delete (window as any).location;
      (window as any).location = { href: '' };

      const error = {
        response: { status: 401 },
        config: {},
        isAxiosError: true,
      };

      await expect(responseErrorHandler!(error)).rejects.toBe(error);
      expect(localStorage.getItem('token')).toBeNull();
      expect(window.location.href).toBe('/login');

      // Restore location
      (window as any).location = originalLocation;
    });

    it('does not redirect or clear token on non-401 errors', async () => {
      localStorage.setItem('token', 'keep-token');
      await import('../../api/client');

      const error = {
        response: { status: 500 },
        config: {},
        isAxiosError: true,
      };

      await expect(responseErrorHandler!(error)).rejects.toBe(error);
      expect(localStorage.getItem('token')).toBe('keep-token');
    });

    it('rejects the error promise even without a response object', async () => {
      await import('../../api/client');

      const error = new Error('Network Error');
      (error as any).config = {};
      (error as any).isAxiosError = true;
      (error as any).response = undefined;

      await expect(responseErrorHandler!(error)).rejects.toBe(error);
    });
  });
});
