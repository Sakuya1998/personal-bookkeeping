import { create } from 'zustand';
import { User, Ledger } from '../api/types';

interface AppState {
  user: User | null;
  token: string | null;
  currentLedger: Ledger | null;
  ledgers: Ledger[];
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setCurrentLedger: (ledger: Ledger | null) => void;
  setLedgers: (ledgers: Ledger[]) => void;
  logout: () => void;
}

export const useAppStore = create<AppState>((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  currentLedger: null,
  ledgers: [],
  setUser: (user) => set({ user }),
  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
    set({ token });
  },
  setCurrentLedger: (ledger) => set({ currentLedger: ledger }),
  setLedgers: (ledgers) => {
    set({ ledgers });
    // Auto-select first ledger if none selected
    const state = useAppStore.getState();
    if (!state.currentLedger && ledgers.length > 0) {
      set({ currentLedger: ledgers[0] });
    }
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null, currentLedger: null });
  },
}));
