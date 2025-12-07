'use client';

import { create } from 'zustand';
import { authAPI, isAuthError, isNetworkError, User, ConvertGuestDTO } from '@/lib/api';

interface AuthStore {
  user: User | null;
  isLoading: boolean;
  connectionError: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, name: string, password: string) => Promise<void>;
  startGuest: () => Promise<void>;
  convertToRegistered: (data: ConvertGuestDTO) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
  setMockUser: (user: User) => void;
  isAnonymous: () => boolean;
}

export const useAuth = create<AuthStore>((set, get) => ({
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

  startGuest: async () => {
    set({ isLoading: true });
    try {
      const response = await authAPI.guest();
      localStorage.setItem('token', response.data.access_token);
      set({ user: response.data.user, isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  convertToRegistered: async (data) => {
    set({ isLoading: true });
    try {
      const response = await authAPI.convert(data);
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
      console.error('[Auth] Session verification failed:', err);

      if (isAuthError(err)) {
        localStorage.removeItem('token');
        set({ user: null, isLoading: false, connectionError: false });
      } else if (isNetworkError(err)) {
        console.warn('[Auth] Network error during session check - keeping token');
        set({ user: null, isLoading: false, connectionError: true });
      } else {
        console.error('[Auth] Unknown error type - clearing token for safety');
        localStorage.removeItem('token');
        set({ user: null, isLoading: false, connectionError: false });
      }
    }
  },

  setMockUser: (user: User) => {
    set({ user, isLoading: false });
  },

  isAnonymous: () => {
    const { user } = get();
    return user?.user_type === 'anonymous';
  },
}));
