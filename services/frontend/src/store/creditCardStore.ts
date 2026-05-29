import { create } from 'zustand';
import { CreditCard, CreditCardFilters, PaymentSchedule } from '../api/creditCards';
import { creditCardsApi } from '../api';

interface CreditCardState {
  creditCards: CreditCard[];
  selectedCard: CreditCard | null;
  filters: CreditCardFilters;
  upcomingPayments: PaymentSchedule[];
  loading: boolean;
  error: string | null;
  
  // Actions
  setCreditCards: (cards: CreditCard[]) => void;
  setSelectedCard: (card: CreditCard | null) => void;
  setFilters: (filters: CreditCardFilters) => void;
  setUpcomingPayments: (payments: PaymentSchedule[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // API Actions
  fetchCreditCards: (filters?: CreditCardFilters) => Promise<void>;
  fetchCardById: (id: string) => Promise<void>;
  createCreditCard: (cardData: Omit<CreditCard, 'id' | 'created_at' | 'updated_at' | 'current_balance' | 'available_credit'>) => Promise<CreditCard>;
  updateCreditCard: (id: string, updates: Partial<CreditCard>) => Promise<CreditCard>;
  deleteCreditCard: (id: string) => Promise<void>;
  softDeleteCreditCard: (id: string) => Promise<void>;
  fetchUpcomingPayments: (daysAhead?: number) => Promise<void>;
  validateCardNumber: (cardNumber: string) => Promise<{ valid: boolean; card_type?: string }>;
  setDefaultCard: (id: string) => Promise<CreditCard>;
  processPayment: (id: string, amount: number, paymentDate?: string) => Promise<CreditCard>;
  getCreditUtilization: () => Promise<{ total_limit: number; total_balance: number; utilization_percentage: number; by_card: Record<string, number> }>;
}

export const useCreditCardStore = create<CreditCardState>((set, get) => ({
  creditCards: [],
  selectedCard: null,
  filters: {},
  upcomingPayments: [],
  loading: false,
  error: null,

  setCreditCards: (creditCards) => set({ creditCards }),
  setSelectedCard: (selectedCard) => set({ selectedCard }),
  setFilters: (filters) => set({ filters }),
  setUpcomingPayments: (upcomingPayments) => set({ upcomingPayments }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  fetchCreditCards: async (filters?: CreditCardFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const creditCards = await creditCardsApi.getAll(actualFilters);
      set({ creditCards, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch credit cards', loading: false });
      throw error;
    }
  },

  fetchCardById: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const card = await creditCardsApi.getById(id);
      set({ selectedCard: card, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch credit card', loading: false });
      throw error;
    }
  },

  createCreditCard: async (cardData) => {
    set({ loading: true, error: null });
    try {
      const card = await creditCardsApi.create(cardData);
      set((state) => ({ 
        creditCards: [card, ...state.creditCards],
        loading: false 
      }));
      return card;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create credit card', loading: false });
      throw error;
    }
  },

  updateCreditCard: async (id: string, updates: Partial<CreditCard>) => {
    set({ loading: true, error: null });
    try {
      const updatedCard = await creditCardsApi.update(id, updates);
      set((state) => ({
        creditCards: state.creditCards.map((c) => 
          c.id === id ? updatedCard : c
        ),
        selectedCard: state.selectedCard?.id === id ? updatedCard : state.selectedCard,
        loading: false,
      }));
      return updatedCard;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update credit card', loading: false });
      throw error;
    }
  },

  deleteCreditCard: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await creditCardsApi.delete(id);
      set((state) => ({
        creditCards: state.creditCards.filter((c) => c.id !== id),
        selectedCard: state.selectedCard?.id === id ? null : state.selectedCard,
        loading: false,
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete credit card', loading: false });
      throw error;
    }
  },

  softDeleteCreditCard: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await creditCardsApi.softDelete(id);
      set((state) => ({
        creditCards: state.creditCards.filter((c) => c.id !== id),
        selectedCard: state.selectedCard?.id === id ? null : state.selectedCard,
        loading: false,
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to soft delete credit card', loading: false });
      throw error;
    }
  },

  fetchUpcomingPayments: async (daysAhead: number = 30) => {
    set({ loading: true, error: null });
    try {
      const payments = await creditCardsApi.getUpcomingPayments(daysAhead);
      set({ upcomingPayments: payments, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch upcoming payments', loading: false });
      throw error;
    }
  },

  validateCardNumber: async (cardNumber: string) => {
    set({ loading: true, error: null });
    try {
      const result = await creditCardsApi.validateCardNumber(cardNumber);
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to validate card number', loading: false });
      throw error;
    }
  },

  setDefaultCard: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const updatedCard = await creditCardsApi.setDefaultCard(id);
      // Update all cards to ensure only one is default
      set((state) => ({
        creditCards: state.creditCards.map((c) => ({
          ...c,
          is_default: c.id === id,
        })),
        selectedCard: state.selectedCard?.id === id ? updatedCard : state.selectedCard,
        loading: false,
      }));
      return updatedCard;
    } catch (error: any) {
      set({ error: error.message || 'Failed to set default card', loading: false });
      throw error;
    }
  },

  processPayment: async (id: string, amount: number, paymentDate?: string) => {
    set({ loading: true, error: null });
    try {
      const updatedCard = await creditCardsApi.processPayment(id, amount, paymentDate);
      set((state) => ({
        creditCards: state.creditCards.map((c) => 
          c.id === id ? updatedCard : c
        ),
        selectedCard: state.selectedCard?.id === id ? updatedCard : state.selectedCard,
        loading: false,
      }));
      return updatedCard;
    } catch (error: any) {
      set({ error: error.message || 'Failed to process payment', loading: false });
      throw error;
    }
  },

  getCreditUtilization: async () => {
    set({ loading: true, error: null });
    try {
      const utilization = await creditCardsApi.getCreditUtilization();
      set({ loading: false });
      return utilization;
    } catch (error: any) {
      set({ error: error.message || 'Failed to get credit utilization', loading: false });
      throw error;
    }
  },
}));