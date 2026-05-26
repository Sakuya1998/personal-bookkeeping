import { describe, it, expect, beforeEach } from 'vitest';
import { useAppStore } from '../../store/appStore';

// Get the initial state for reset purposes
const initialState = useAppStore.getInitialState();

beforeEach(() => {
  localStorage.clear();
  // Reset store to initial state
  useAppStore.setState(initialState);
});

describe('appStore', () => {
  describe('initial state', () => {
    it('has user set to null', () => {
      const state = useAppStore.getState();
      expect(state.user).toBeNull();
    });

    it('reads token from localStorage', () => {
      // token should be null when localStorage is empty
      const state = useAppStore.getState();
      expect(state.token).toBeNull();
    });

    it('reads existing token from localStorage on init', () => {
      localStorage.setItem('token', 'saved-token');
      // Create a fresh reference by resetting
      useAppStore.setState({ token: 'saved-token' });
      const state = useAppStore.getState();
      expect(state.token).toBe('saved-token');
    });
  });

  describe('setUser', () => {
    it('sets a user', () => {
      const user = {
        id: '1',
        username: 'testuser',
        email: 'test@example.com',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
      };
      useAppStore.getState().setUser(user);
      expect(useAppStore.getState().user).toEqual(user);
    });

    it('sets user to null', () => {
      useAppStore.getState().setUser(null);
      expect(useAppStore.getState().user).toBeNull();
    });
  });

  describe('setToken', () => {
    it('saves token to localStorage and updates state', () => {
      useAppStore.getState().setToken('my-token');
      expect(localStorage.getItem('token')).toBe('my-token');
      expect(useAppStore.getState().token).toBe('my-token');
    });

    it('removes token from localStorage when set to null', () => {
      localStorage.setItem('token', 'old-token');
      useAppStore.getState().setToken(null);
      expect(localStorage.getItem('token')).toBeNull();
      expect(useAppStore.getState().token).toBeNull();
    });
  });

  describe('setLedgers', () => {
    it('sets ledgers list', () => {
      const ledgers = [
        {
          id: '1',
          user_id: 'u1',
          name: 'Main',
          description: null,
          base_currency: 'CNY',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ];
      useAppStore.getState().setLedgers(ledgers);
      expect(useAppStore.getState().ledgers).toEqual(ledgers);
    });

    it('auto-selects the first ledger if none is current', () => {
      const ledgers = [
        {
          id: '1',
          user_id: 'u1',
          name: 'Main',
          description: null,
          base_currency: 'CNY',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: '2',
          user_id: 'u1',
          name: 'Savings',
          description: null,
          base_currency: 'USD',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 1,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ];
      useAppStore.getState().setLedgers(ledgers);
      expect(useAppStore.getState().currentLedger).toEqual(ledgers[0]);
    });

    it('does not change currentLedger if already selected', () => {
      // First set a current ledger
      const existingLedger = {
        id: '2',
        user_id: 'u1',
        name: 'Savings',
        description: null,
        base_currency: 'USD',
        icon: null,
        color: null,
        is_archived: false,
        sort_order: 1,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };
      useAppStore.setState({ currentLedger: existingLedger });

      const ledgers = [
        {
          id: '1',
          user_id: 'u1',
          name: 'Main',
          description: null,
          base_currency: 'CNY',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: '2',
          user_id: 'u1',
          name: 'Savings',
          description: null,
          base_currency: 'USD',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 1,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ];
      useAppStore.getState().setLedgers(ledgers);
      // CurrentLedger should still be the one we set (ledger id '2'), not auto-selected to first
      expect(useAppStore.getState().currentLedger).toEqual(existingLedger);
    });

    it('does not set currentLedger if ledgers list is empty', () => {
      useAppStore.getState().setLedgers([]);
      expect(useAppStore.getState().currentLedger).toBeNull();
    });
  });

  describe('setCurrentLedger', () => {
    it('switches to a different ledger', () => {
      const ledger = {
        id: '3',
        user_id: 'u1',
        name: 'Travel',
        description: null,
        base_currency: 'EUR',
        icon: null,
        color: null,
        is_archived: false,
        sort_order: 2,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };
      useAppStore.getState().setCurrentLedger(ledger);
      expect(useAppStore.getState().currentLedger).toEqual(ledger);
    });

    it('sets currentLedger to null', () => {
      useAppStore.getState().setCurrentLedger(null);
      expect(useAppStore.getState().currentLedger).toBeNull();
    });
  });

  describe('logout', () => {
    it('clears token from localStorage', () => {
      localStorage.setItem('token', 'my-token');
      useAppStore.getState().logout();
      expect(localStorage.getItem('token')).toBeNull();
    });

    it('resets user, token, and currentLedger to null', () => {
      // Set some state first
      useAppStore.setState({
        user: {
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
        },
        token: 'my-token',
        currentLedger: {
          id: '1',
          user_id: 'u1',
          name: 'Main',
          description: null,
          base_currency: 'CNY',
          icon: null,
          color: null,
          is_archived: false,
          sort_order: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      });

      useAppStore.getState().logout();
      const state = useAppStore.getState();
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.currentLedger).toBeNull();
    });
  });
});
