import { create } from 'zustand';
import { ExchangeRate, CurrencyConversion, SupportedCurrency } from '../api/exchangeRates';
import { exchangeRatesApi } from '../api';

interface ExchangeRateState {
  exchangeRates: ExchangeRate[];
  latestRates: ExchangeRate[];
  supportedCurrencies: SupportedCurrency[];
  selectedRate: ExchangeRate | null;
  conversionResult: CurrencyConversion | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  setExchangeRates: (rates: ExchangeRate[]) => void;
  setLatestRates: (rates: ExchangeRate[]) => void;
  setSupportedCurrencies: (currencies: SupportedCurrency[]) => void;
  setSelectedRate: (rate: ExchangeRate | null) => void;
  setConversionResult: (result: CurrencyConversion | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // API Actions
  fetchLatestRates: () => Promise<void>;
  fetchAllRates: () => Promise<void>;
  fetchRateById: (id: string) => Promise<void>;
  fetchRateByPair: (baseCurrency: string, targetCurrency: string) => Promise<void>;
  fetchRateByDate: (baseCurrency: string, targetCurrency: string, date: string) => Promise<void>;
  fetchRatesByDateRange: (baseCurrency: string, targetCurrency: string, startDate: string, endDate: string) => Promise<void>;
  fetchSupportedCurrencies: () => Promise<void>;
  createExchangeRate: (rate: Omit<ExchangeRate, 'id' | 'created_at' | 'updated_at'>) => Promise<ExchangeRate>;
  updateExchangeRate: (id: string, updates: Partial<ExchangeRate>) => Promise<ExchangeRate>;
  deleteExchangeRate: (id: string) => Promise<void>;
  convertCurrency: (amount: number, fromCurrency: string, toCurrency: string, date?: string) => Promise<CurrencyConversion>;
  updateRatesFromAPI: () => Promise<{ updated: number; failed: number }>;
  getRateWithFallback: (baseCurrency: string, targetCurrency: string) => Promise<ExchangeRate>;
  getCurrencyStats: () => Promise<{
    total_currencies: number;
    total_rates: number;
    last_updated: string;
    most_active: Array<{ currency: string; rate_count: number }>;
  }>;
  validateCurrency: (currencyCode: string) => Promise<{ valid: boolean; currency?: SupportedCurrency }>;
  calculateFees: (amount: number, fromCurrency: string, toCurrency: string) => Promise<{
    amount: number;
    converted_amount: number;
    rate: number;
    fee_percentage: number;
    fee_amount: number;
    total_amount: number;
  }>;
}

export const useExchangeRateStore = create<ExchangeRateState>((set, get) => ({
  exchangeRates: [],
  latestRates: [],
  supportedCurrencies: [],
  selectedRate: null,
  conversionResult: null,
  loading: false,
  error: null,

  setExchangeRates: (exchangeRates) => set({ exchangeRates }),
  setLatestRates: (latestRates) => set({ latestRates }),
  setSupportedCurrencies: (supportedCurrencies) => set({ supportedCurrencies }),
  setSelectedRate: (selectedRate) => set({ selectedRate }),
  setConversionResult: (conversionResult) => set({ conversionResult }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  fetchLatestRates: async () => {
    set({ loading: true, error: null });
    try {
      const latestRates = await exchangeRatesApi.getAllLatest();
      set({ latestRates, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch latest rates', loading: false });
      throw error;
    }
  },

  fetchAllRates: async () => {
    set({ loading: true, error: null });
    try {
      const rates = await exchangeRatesApi.getAllLatest();
      set({ exchangeRates: rates, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch exchange rates', loading: false });
      throw error;
    }
  },

  fetchRateById: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const rate = await exchangeRatesApi.getById(id);
      set({ selectedRate: rate, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch exchange rate', loading: false });
      throw error;
    }
  },

  fetchRateByPair: async (baseCurrency: string, targetCurrency: string) => {
    set({ loading: true, error: null });
    try {
      const rate = await exchangeRatesApi.getLatest(baseCurrency, targetCurrency);
      set({ selectedRate: rate, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch exchange rate pair', loading: false });
      throw error;
    }
  },

  fetchRateByDate: async (baseCurrency: string, targetCurrency: string, date: string) => {
    set({ loading: true, error: null });
    try {
      const rate = await exchangeRatesApi.getByDate(baseCurrency, targetCurrency, date);
      set({ selectedRate: rate, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch exchange rate by date', loading: false });
      throw error;
    }
  },

  fetchRatesByDateRange: async (baseCurrency: string, targetCurrency: string, startDate: string, endDate: string) => {
    set({ loading: true, error: null });
    try {
      const rates = await exchangeRatesApi.getByDateRange(baseCurrency, targetCurrency, startDate, endDate);
      set({ exchangeRates: rates, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch exchange rates by date range', loading: false });
      throw error;
    }
  },

  fetchSupportedCurrencies: async () => {
    set({ loading: true, error: null });
    try {
      const currencies = await exchangeRatesApi.getSupportedCurrencies();
      set({ supportedCurrencies: currencies, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch supported currencies', loading: false });
      throw error;
    }
  },

  createExchangeRate: async (rateData) => {
    set({ loading: true, error: null });
    try {
      const rate = await exchangeRatesApi.create(rateData);
      set((state) => ({ 
        exchangeRates: [rate, ...state.exchangeRates],
        latestRates: [rate, ...state.latestRates.filter(r => 
          !(r.base_currency === rate.base_currency && r.target_currency === rate.target_currency)
        )],
        loading: false 
      }));
      return rate;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create exchange rate', loading: false });
      throw error;
    }
  },

  updateExchangeRate: async (id: string, updates: Partial<ExchangeRate>) => {
    set({ loading: true, error: null });
    try {
      const updatedRate = await exchangeRatesApi.update(id, updates);
      set((state) => ({
        exchangeRates: state.exchangeRates.map((r) => 
          r.id === id ? updatedRate : r
        ),
        latestRates: state.latestRates.map((r) =>
          r.id === id ? updatedRate : r
        ),
        selectedRate: state.selectedRate?.id === id ? updatedRate : state.selectedRate,
        loading: false,
      }));
      return updatedRate;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update exchange rate', loading: false });
      throw error;
    }
  },

  deleteExchangeRate: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await exchangeRatesApi.delete(id);
      set((state) => ({
        exchangeRates: state.exchangeRates.filter((r) => r.id !== id),
        latestRates: state.latestRates.filter((r) => r.id !== id),
        selectedRate: state.selectedRate?.id === id ? null : state.selectedRate,
        loading: false,
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete exchange rate', loading: false });
      throw error;
    }
  },

  convertCurrency: async (amount: number, fromCurrency: string, toCurrency: string, date?: string) => {
    set({ loading: true, error: null });
    try {
      const result = await exchangeRatesApi.convertAmount(amount, fromCurrency, toCurrency, date);
      set({ conversionResult: result, loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to convert currency', loading: false });
      throw error;
    }
  },

  updateRatesFromAPI: async () => {
    set({ loading: true, error: null });
    try {
      const result = await exchangeRatesApi.updateRatesFromAPI();
      // Refresh rates after update
      await get().fetchLatestRates();
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update rates from API', loading: false });
      throw error;
    }
  },

  getRateWithFallback: async (baseCurrency: string, targetCurrency: string) => {
    set({ loading: true, error: null });
    try {
      const rate = await exchangeRatesApi.getRateWithFallback(baseCurrency, targetCurrency);
      set({ selectedRate: rate, loading: false });
      return rate;
    } catch (error: any) {
      set({ error: error.message || 'Failed to get rate with fallback', loading: false });
      throw error;
    }
  },

  getCurrencyStats: async () => {
    set({ loading: true, error: null });
    try {
      const stats = await exchangeRatesApi.getCurrencyStats();
      set({ loading: false });
      return stats;
    } catch (error: any) {
      set({ error: error.message || 'Failed to get currency stats', loading: false });
      throw error;
    }
  },

  validateCurrency: async (currencyCode: string) => {
    set({ loading: true, error: null });
    try {
      const result = await exchangeRatesApi.validateCurrency(currencyCode);
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to validate currency', loading: false });
      throw error;
    }
  },

  calculateFees: async (amount: number, fromCurrency: string, toCurrency: string) => {
    set({ loading: true, error: null });
    try {
      const fees = await exchangeRatesApi.calculateFees(amount, fromCurrency, toCurrency);
      set({ loading: false });
      return fees;
    } catch (error: any) {
      set({ error: error.message || 'Failed to calculate fees', loading: false });
      throw error;
    }
  },
}));