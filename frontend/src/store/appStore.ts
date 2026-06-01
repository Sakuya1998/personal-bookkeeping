import { create } from 'zustand';
import { User, Ledger, MemberRole } from '../api/types';

interface AppState {
  user: User | null;
  token: string | null;
  currentLedger: Ledger | null;
  currentRole: MemberRole | null;
  ledgers: Ledger[];
  setCurrentRole: (role: MemberRole | null) => void;
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
  currentRole: null,
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
  setCurrentRole: (role: MemberRole | null) => set({ currentRole: role }),
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
    set({ user: null, token: null, currentLedger: null, currentRole: null });
  },
}));
