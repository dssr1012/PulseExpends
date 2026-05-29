import api from './client';

export interface CreditCard {
  id: string;
  user_id: string;
  circle_id?: string;
  bank_name: string;
  card_type: 'visa' | 'mastercard' | 'amex' | 'discover' | 'other';
  last_four: string;
  cardholder_name: string;
  expiration_month: number;
  expiration_year: number;
  billing_address?: string;
  credit_limit?: number;
  current_balance: number;
  available_credit: number;
  payment_due_date?: string;
  minimum_payment?: number;
  statement_balance?: number;
  is_default: boolean;
  is_active: boolean;
  deleted_at?: string;
  created_at: string;
  updated_at: string;
}

export interface PaymentSchedule {
  due_date: string;
  minimum_payment: number;
  statement_balance: number;
  is_paid: boolean;
  paid_at?: string;
}

export interface CreditCardFilters {
  bank_name?: string;
  card_type?: string;
  is_default?: boolean;
  is_active?: boolean;
  circle_id?: string;
}

export const creditCardApi = {
  // Create a new credit card (NO PAN/CVV storage - only last 4 digits)
  create: async (card: Omit<CreditCard, 'id' | 'created_at' | 'updated_at' | 'current_balance' | 'available_credit'>) => {
    const response = await api.post<CreditCard>('/api/credit-cards', card);
    return response.data;
  },

  // Get credit card by ID
  getById: async (id: string) => {
    const response = await api.get<CreditCard>(`/api/credit-cards/${id}`);
    return response.data;
  },

  // Get all credit cards for user
  getAll: async (filters?: CreditCardFilters) => {
    const response = await api.get<CreditCard[]>('/api/credit-cards', { params: filters });
    return response.data;
  },

  // Update credit card
  update: async (id: string, updates: Partial<CreditCard>) => {
    const response = await api.put<CreditCard>(`/api/credit-cards/${id}`, updates);
    return response.data;
  },

  // Delete credit card (hard delete)
  delete: async (id: string) => {
    const response = await api.delete(`/api/credit-cards/${id}`);
    return response.data;
  },

  // Soft delete credit card
  softDelete: async (id: string) => {
    const response = await api.delete(`/api/credit-cards/${id}/soft`);
    return response.data;
  },

  // Get credit cards by bank
  getByBank: async (bank_name: string) => {
    const response = await api.get<CreditCard[]>(`/api/credit-cards/bank/${bank_name}`);
    return response.data;
  },

  // Get credit cards by last four digits
  getByLastFour: async (last_four: string) => {
    const response = await api.get<CreditCard[]>(`/api/credit-cards/last-four/${last_four}`);
    return response.data;
  },

  // Get active credit cards
  getActiveCards: async () => {
    const response = await api.get<CreditCard[]>('/api/credit-cards/active');
    return response.data;
  },

  // Get default credit card
  getDefaultCard: async () => {
    const response = await api.get<CreditCard | null>('/api/credit-cards/default');
    return response.data;
  },

  // Update credit card balance
  updateBalance: async (id: string, amount: number, transaction_type: 'charge' | 'payment' | 'credit') => {
    const response = await api.patch<CreditCard>(`/api/credit-cards/${id}/balance`, { amount, transaction_type });
    return response.data;
  },

  // Get upcoming payments
  getUpcomingPayments: async (days_ahead: number = 30) => {
    const response = await api.get<PaymentSchedule[]>(`/api/credit-cards/payments/upcoming?days_ahead=${days_ahead}`);
    return response.data;
  },

  // Get statement period
  getStatementPeriod: async (id: string, year: number, month: number) => {
    const response = await api.get<{
      start_date: string;
      end_date: string;
      opening_balance: number;
      closing_balance: number;
      total_charges: number;
      total_payments: number;
      transactions: any[];
    }>(`/api/credit-cards/${id}/statement/${year}/${month}`);
    return response.data;
  },

  // Validate credit card number (Luhn algorithm)
  validateCardNumber: async (card_number: string) => {
    const response = await api.post<{ valid: boolean; card_type?: string }>('/api/credit-cards/validate', { card_number });
    return response.data;
  },

  // Calculate minimum payment
  calculateMinimumPayment: async (id: string, balance?: number) => {
    const response = await api.get<{ minimum_payment: number; due_date: string }>(
      `/api/credit-cards/${id}/calculate-minimum-payment${balance ? `?balance=${balance}` : ''}`
    );
    return response.data;
  },

  // Get credit utilization
  getCreditUtilization: async () => {
    const response = await api.get<{ total_limit: number; total_balance: number; utilization_percentage: number; by_card: Record<string, number> }>(
      '/api/credit-cards/utilization'
    );
    return response.data;
  },

  // Set default credit card
  setDefaultCard: async (id: string) => {
    const response = await api.patch<CreditCard>(`/api/credit-cards/${id}/set-default`);
    return response.data;
  },

  // Get payment history
  getPaymentHistory: async (id: string, limit: number = 50) => {
    const response = await api.get<any[]>(`/api/credit-cards/${id}/payment-history?limit=${limit}`);
    return response.data;
  },

  // Process payment
  processPayment: async (id: string, amount: number, payment_date?: string) => {
    const response = await api.post<CreditCard>(`/api/credit-cards/${id}/process-payment`, { amount, payment_date });
    return response.data;
  },

  // Get expired cards
  getExpiredCards: async () => {
    const response = await api.get<CreditCard[]>('/api/credit-cards/expired');
    return response.data;
  },

  // Renew card (generate new last four)
  renewCard: async (id: string) => {
    const response = await api.post<CreditCard>(`/api/credit-cards/${id}/renew`);
    return response.data;
  },
};