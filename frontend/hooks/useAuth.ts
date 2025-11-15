'use client';

import { create } from 'zustand';
import { authAPI } from '@/lib/api';

interface User {
  id: string;
  email: string;
  name: string;
}

interface AuthStore {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, name: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
}

export const useAuth = create<AuthStore>((set) => ({
  user: null,
  isLoading: false,

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

    set({ isLoading: true });
    try {
      const response = await authAPI.me();
      set({ user: response.data, isLoading: false });
    } catch (error) {
      localStorage.removeItem('token');
      set({ user: null, isLoading: false });
    }
  },
}));
