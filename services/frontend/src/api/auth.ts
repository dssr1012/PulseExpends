import api from './client';
import type { AuthResponse, LoginRequest, RegisterRequest, User } from '../types';

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await api.post<AuthResponse>('/api/auth/login', data);
  return res.data;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await api.post<AuthResponse>('/api/auth/register', data);
  return res.data;
}

export async function getProfile(): Promise<User> {
  const res = await api.get<User>('/api/auth/profile');
  return res.data;
}

export async function logout(): Promise<void> {
  try {
    await api.post('/api/auth/logout');
  } finally {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('current_user');
  }
}
