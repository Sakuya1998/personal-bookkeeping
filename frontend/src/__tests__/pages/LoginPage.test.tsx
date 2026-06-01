import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import React from 'react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

const mockPost = vi.fn();

vi.mock('../../api/client', () => ({
  default: {
    post: mockPost,
  },
}));

const mockSetToken = vi.fn();
const mockSetUser = vi.fn();

vi.mock('../../store/appStore', () => ({
  useAppStore: vi.fn(() => ({
    setToken: mockSetToken,
    setUser: mockSetUser,
  })),
}));

function makeAuthResponse() {
  return {
    data: {
      code: 200,
      message: 'ok',
      data: {
        token: 't',
        user: { id: 'u1', username: 'u', email: 'u@example.com' },
      },
    },
  };
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('redirects to next after login success', async () => {
    const mod = await import('../../pages/LoginPage');
    const LoginPage = mod.default;

    mockPost.mockResolvedValue(makeAuthResponse());

    render(
      <MemoryRouter initialEntries={['/login?next=/transactions']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/transactions" element={<div>TX</div>} />
          <Route path="/" element={<div>HOME</div>} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText('auth.username'), { target: { value: 'u' } });
    fireEvent.change(screen.getByLabelText('auth.password'), { target: { value: 'p' } });
    fireEvent.click(screen.getByRole('button', { name: 'auth.login' }));

    await waitFor(() => {
      expect(screen.getByText('TX')).toBeInTheDocument();
    });
  });

  it('falls back to / when next is not a safe path', async () => {
    const mod = await import('../../pages/LoginPage');
    const LoginPage = mod.default;

    mockPost.mockResolvedValue(makeAuthResponse());

    render(
      <MemoryRouter initialEntries={['/login?next=https://evil.com']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/transactions" element={<div>TX</div>} />
          <Route path="/" element={<div>HOME</div>} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText('auth.username'), { target: { value: 'u' } });
    fireEvent.change(screen.getByLabelText('auth.password'), { target: { value: 'p' } });
    fireEvent.click(screen.getByRole('button', { name: 'auth.login' }));

    await waitFor(() => {
      expect(screen.getByText('HOME')).toBeInTheDocument();
    });
  });
});
