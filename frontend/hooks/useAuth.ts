'use client';

import { create } from 'zustand';
import { authAPI, isAuthError, isNetworkError } from '@/lib/api';

interface User {
  id: string;
  email: string;
  name: string;
}

interface AuthStore {
  user: User | null;
  isLoading: boolean;
  connectionError: boolean; // True when network error prevented auth check
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, name: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
  setMockUser: (user: User) => void;
}

export const useAuth = create<AuthStore>((set) => ({
  user: null,
  isLoading: false,
  connectionError: false,

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const response = await authAPI.login({ email, password });
      localStorage.setItem('token', response.data.access_token);
      set({ user: response.data.user, isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  register: async (email, name, password) => {
    set({ isLoading: true });
    try {
      const response = await authAPI.register({ email, name, password });
      localStorage.setItem('token', response.data.access_token);
      set({ user: response.data.user, isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null });
  },

  checkAuth: async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      set({ user: null, isLoading: false });
      return;
    }

    set({ isLoading: true, connectionError: false });
    try {
      const response = await authAPI.me();
      set({ user: response.data, isLoading: false, connectionError: false });
    } catch (err: unknown) {
      // Always log the error for debugging
      console.error('[Auth] Session verification failed:', err);

      // Only clear token for authentication errors (401, 403)
      // For network errors, keep the token and let user retry
      if (isAuthError(err)) {
        localStorage.removeItem('token');
        set({ user: null, isLoading: false, connectionError: false });
      } else if (isNetworkError(err)) {
        // Network error - keep token and set connectionError flag
        // User can retry when connection is restored
        console.warn('[Auth] Network error during session check - keeping token');
        set({ user: null, isLoading: false, connectionError: true });
      } else {
        // Unknown error - clear token to be safe, notify user
        console.error('[Auth] Unknown error type - clearing token for safety');
        localStorage.removeItem('token');
        set({ user: null, isLoading: false, connectionError: false });
      }
    }
  },

  setMockUser: (user: User) => {
    set({ user, isLoading: false });
  },
}));
