import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { App as AntApp } from 'antd';
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom';
import AuthEventBridge from '../../components/AuthEventBridge';

const LocationDisplay: React.FC = () => {
  const location = useLocation();
  return <div data-testid="loc">{`${location.pathname}${location.search}`}</div>;
};

describe('AuthEventBridge', () => {
  it('navigates to /login with encoded next when auth:unauthorized is dispatched', async () => {
    render(
      <AntApp>
        <MemoryRouter initialEntries={['/transactions?x=1']}>
          <AuthEventBridge />
          <Routes>
            <Route path="/login" element={<LocationDisplay />} />
            <Route path="*" element={<LocationDisplay />} />
          </Routes>
        </MemoryRouter>
      </AntApp>,
    );

    expect(screen.getByTestId('loc')).toHaveTextContent('/transactions?x=1');

    window.dispatchEvent(new CustomEvent('auth:unauthorized', { detail: { next: '/transactions?x=1' } }));

    await waitFor(() => {
      expect(screen.getByTestId('loc')).toHaveTextContent('/login?next=%2Ftransactions%3Fx%3D1');
    });
  });
});

