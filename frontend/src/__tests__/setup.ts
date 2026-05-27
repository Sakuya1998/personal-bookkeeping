import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock window.matchMedia for antd responsive components (Grid, etc.)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock ResizeObserver for antd components (Modal, etc.)
// Use a real class (not vi.fn()) so `new ResizeObserver(...)` works
Object.defineProperty(window, 'ResizeObserver', {
  writable: true,
  value: class MockResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  },
});
