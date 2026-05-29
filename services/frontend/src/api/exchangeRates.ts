import api from './client';

export interface ExchangeRate {
  id: string;
  base_currency: string;
  target_currency: string;
  rate: number;
  source: 'api' | 'manual' | 'fallback';
  effective_date: string;
  created_at: string;
  updated_at: string;
}

export interface CurrencyConversion {
  amount: number;
  from_currency: string;
  to_currency: string;
  rate: number;
  converted_amount: number;
  timestamp: string;
}

export interface SupportedCurrency {
  code: string;
  name: string;
  symbol: string;
  is_active: boolean;
}

export const exchangeRateApi = {
  // Create a new exchange rate
  create: async (rate: Omit<ExchangeRate, 'id' | 'created_at' | 'updated_at'>) => {
    const response = await api.post<ExchangeRate>('/api/exchange-rates', rate);
    return response.data;
  },

  // Get exchange rate by ID
  getById: async (id: string) => {
    const response = await api.get<ExchangeRate>(`/api/exchange-rates/${id}`);
    return response.data;
  },

  // Get latest exchange rate for currency pair
  getLatest: async (base_currency: string, target_currency: string) => {
    const response = await api.get<ExchangeRate>(`/api/exchange-rates/latest/${base_currency}/${target_currency}`);
    return response.data;
  },

  // Get exchange rate for specific date
  getByDate: async (base_currency: string, target_currency: string, date: string) => {
    const response = await api.get<ExchangeRate>(`/api/exchange-rates/date/${base_currency}/${target_currency}/${date}`);
    return response.data;
  },

  // Get exchange rates for date range
  getByDateRange: async (base_currency: string, target_currency: string, start_date: string, end_date: string) => {
    const response = await api.get<ExchangeRate[]>(
      `/api/exchange-rates/range/${base_currency}/${target_currency}?start_date=${start_date}&end_date=${end_date}`
    );
    return response.data;
  },

  // Get all latest exchange rates
  getAllLatest: async () => {
    const response = await api.get<ExchangeRate[]>('/api/exchange-rates/latest');
    return response.data;
  },

  // Update exchange rate
  update: async (id: string, updates: Partial<ExchangeRate>) => {
    const response = await api.put<ExchangeRate>(`/api/exchange-rates/${id}`, updates);
    return response.data;
  },

  // Delete exchange rate
  delete: async (id: string) => {
    const response = await api.delete(`/api/exchange-rates/${id}`);
    return response.data;
  },

  // Convert currency amount
  convertAmount: async (amount: number, from_currency: string, to_currency: string, date?: string) => {
    const response = await api.post<CurrencyConversion>('/api/exchange-rates/convert', {
      amount,
      from_currency,
      to_currency,
      date,
    });
    return response.data;
  },

  // Get supported currencies
  getSupportedCurrencies: async () => {
    const response = await api.get<SupportedCurrency[]>('/api/exchange-rates/currencies');
    return response.data;
  },

  // Update rates from external API
  updateRatesFromAPI: async () => {
    const response = await api.post<{ updated: number; failed: number }>('/api/exchange-rates/update-from-api');
    return response.data;
  },

  // Get rate with fallback (tries direct rate, then inverse, then manual)
  getRateWithFallback: async (base_currency: string, target_currency: string) => {
    const response = await api.get<ExchangeRate>(`/api/exchange-rates/with-fallback/${base_currency}/${target_currency}`);
    return response.data;
  },

  // Get currency statistics
  getCurrencyStats: async () => {
    const response = await api.get<{
      total_currencies: number;
      total_rates: number;
      last_updated: string;
      most_active: Array<{ currency: string; rate_count: number }>;
    }>('/api/exchange-rates/stats');
    return response.data;
  },

  // Bulk create exchange rates
  createBatch: async (rates: Omit<ExchangeRate, 'id' | 'created_at' | 'updated_at'>[]) => {
    const response = await api.post<ExchangeRate[]>('/api/exchange-rates/batch', rates);
    return response.data;
  },

  // Get rate history for currency pair
  getRateHistory: async (base_currency: string, target_currency: string, limit: number = 100) => {
    const response = await api.get<ExchangeRate[]>(`/api/exchange-rates/history/${base_currency}/${target_currency}?limit=${limit}`);
    return response.data;
  },

  // Validate currency code
  validateCurrency: async (currency_code: string) => {
    const response = await api.get<{ valid: boolean; currency?: SupportedCurrency }>(`/api/exchange-rates/validate/${currency_code}`);
    return response.data;
  },

  // Get conversion history for user
  getConversionHistory: async (limit: number = 50) => {
    const response = await api.get<CurrencyConversion[]>(`/api/exchange-rates/conversions/history?limit=${limit}`);
    return response.data;
  },

  // Calculate conversion fees
  calculateFees: async (amount: number, from_currency: string, to_currency: string) => {
    const response = await api.get<{
      amount: number;
      converted_amount: number;
      rate: number;
      fee_percentage: number;
      fee_amount: number;
      total_amount: number;
    }>(`/api/exchange-rates/calculate-fees?amount=${amount}&from=${from_currency}&to=${to_currency}`);
    return response.data;
  },
};