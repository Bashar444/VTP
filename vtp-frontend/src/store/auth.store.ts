import { create } from 'zustand';
import type { User } from '@/types/api';

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  setAuth: (user: User, token: string, refreshToken: string) => void;
  clearAuth: () => void;
  setLoading: (loading: boolean) => void;
  setUser: (user: User) => void;
  initFromStorage: () => void;
}

// Helper to safely access localStorage
const getFromStorage = (key: string): string | null => {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  isLoading: false,

  initFromStorage: () => {
    const token = getFromStorage('authToken');
    const refreshToken = getFromStorage('refreshToken');
    const userStr = getFromStorage('user');
    
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        set({
          user,
          token,
          refreshToken: refreshToken || null,
          isAuthenticated: true,
        });
      } catch {
        // Invalid stored data, clear it
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('user');
      }
    }
  },

  setAuth: (user: User, token: string, refreshToken: string) => {
    set({
      user,
      token,
      refreshToken,
      isAuthenticated: true,
    });
    // Persist to localStorage
    if (typeof window !== 'undefined') {
      localStorage.setItem('authToken', token);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('user', JSON.stringify(user));
    }
  },

  clearAuth: () => {
    set({
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
    });
    if (typeof window !== 'undefined') {
      localStorage.removeItem('authToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
    }
  },

  setLoading: (loading: boolean) => set({ isLoading: loading }),

  setUser: (user: User) => set({ user }),
}));
