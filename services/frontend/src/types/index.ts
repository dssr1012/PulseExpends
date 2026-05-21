export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  display_name: string;
  avatar_url: string | null;
  preferences: UserPreferences;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UserPreferences {
  currency: string;
  language: string;
  theme: 'light' | 'dark';
  date_format: string;
}

export interface UserSession {
  id: string;
  user_id: string;
  token: string;
  expires_at: string;
  last_used_at: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  display_name?: string;
}

export interface Circle {
  id: string;
  name: string;
  description: string;
  type: 'personal' | 'family' | 'group';
  settings: CircleSettings;
  member_count: number;
  role: 'owner' | 'admin' | 'member';
  created_at: string;
  updated_at: string;
}

export interface CircleSettings {
  allow_member_invites: boolean;
  require_approval: boolean;
  default_currency: string;
  monthly_budget: number | null;
}

export interface CircleMember {
  id: string;
  circle_id: string;
  user_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
}

export interface Transaction {
  id: string;
  circle_id: string;
  user_id: string;
  type: 'income' | 'expense' | 'transfer';
  amount: number;
  currency: string;
  category: string;
  description: string;
  date: string;
  created_at: string;
}

export interface ApiError {
  error: string;
  message: string;
  status?: number;
}
