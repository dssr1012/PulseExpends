import { create } from 'zustand';
import type { User } from '../types';
import * as authApi from '../api/auth';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  isAuthenticated: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, firstName: string, lastName: string, displayName?: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isLoading: false,
  error: null,
  isAuthenticated: false,

  hydrate: () => {
    const token = localStorage.getItem('jwt_token');
    const userJson = localStorage.getItem('current_user');
    if (token && userJson) {
      try {
        const user = JSON.parse(userJson) as User;
        set({ token, user, isAuthenticated: true });
      } catch {
        localStorage.removeItem('jwt_token');
        localStorage.removeItem('current_user');
      }
    }
  },

  login: async (email, password) => {
    set({ isLoading: true, error: null });
    try {
      const res = await authApi.login({ email, password });
      localStorage.setItem('jwt_token', res.token);
      localStorage.setItem('current_user', JSON.stringify(res.user));
      set({ token: res.token, user: res.user, isAuthenticated: true, isLoading: false });
    } catch (err: any) {
      const message = err.response?.data?.error || err.response?.data?.message || 'Login failed';
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  register: async (email, password, firstName, lastName, displayName) => {
    set({ isLoading: true, error: null });
    try {
      const res = await authApi.register({
        email,
        password,
        first_name: firstName,
        last_name: lastName,
        display_name: displayName,
      });
      localStorage.setItem('jwt_token', res.token);
      localStorage.setItem('current_user', JSON.stringify(res.user));
      set({ token: res.token, user: res.user, isAuthenticated: true, isLoading: false });
    } catch (err: any) {
      const message = err.response?.data?.error || err.response?.data?.message || 'Registration failed';
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } finally {
      localStorage.removeItem('jwt_token');
      localStorage.removeItem('current_user');
      set({ user: null, token: null, isAuthenticated: false });
    }
  },

  clearError: () => set({ error: null }),
}));
