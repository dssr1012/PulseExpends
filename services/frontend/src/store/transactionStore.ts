import { create } from 'zustand';
import { Transaction, TransactionFilters, TransactionSummary, TransactionAnomalyStats } from '../api/transactions';
import { transactionsApi } from '../api';

interface TransactionState {
  transactions: Transaction[];
  selectedTransaction: Transaction | null;
  filters: TransactionFilters;
  summary: TransactionSummary | null;
  anomalyStats: TransactionAnomalyStats | null;
  loading: boolean;
  error: string | null;
  
  // Actions
  setTransactions: (transactions: Transaction[]) => void;
  setSelectedTransaction: (transaction: Transaction | null) => void;
  setFilters: (filters: TransactionFilters) => void;
  setSummary: (summary: TransactionSummary | null) => void;
  setAnomalyStats: (stats: TransactionAnomalyStats | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // API Actions
  fetchTransactions: (filters?: TransactionFilters) => Promise<void>;
  fetchTransactionById: (id: string) => Promise<void>;
  createTransaction: (transaction: Omit<Transaction, 'id' | 'created_at' | 'updated_at'>) => Promise<Transaction>;
  updateTransaction: (id: string, updates: Partial<Transaction>) => Promise<Transaction>;
  deleteTransaction: (id: string) => Promise<void>;
  fetchSummary: (filters?: TransactionFilters) => Promise<void>;
  fetchAnomalyStats: () => Promise<void>;
  exportToCSV: (filters?: TransactionFilters) => Promise<Blob>;
  importFromCSV: (file: File) => Promise<{ imported: number; skipped: number; errors: string[] }>;
}

export const useTransactionStore = create<TransactionState>((set, get) => ({
  transactions: [],
  selectedTransaction: null,
  filters: {},
  summary: null,
  anomalyStats: null,
  loading: false,
  error: null,

  setTransactions: (transactions) => set({ transactions }),
  setSelectedTransaction: (transaction) => set({ selectedTransaction: transaction }),
  setFilters: (filters) => set({ filters }),
  setSummary: (summary) => set({ summary }),
  setAnomalyStats: (anomalyStats) => set({ anomalyStats }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  fetchTransactions: async (filters?: TransactionFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const transactions = await transactionsApi.getAll(actualFilters);
      set({ transactions, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch transactions', loading: false });
      throw error;
    }
  },

  fetchTransactionById: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const transaction = await transactionsApi.getById(id);
      set({ selectedTransaction: transaction, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch transaction', loading: false });
      throw error;
    }
  },

  createTransaction: async (transactionData) => {
    set({ loading: true, error: null });
    try {
      const transaction = await transactionsApi.create(transactionData);
      set((state) => ({ 
        transactions: [transaction, ...state.transactions],
        loading: false 
      }));
      return transaction;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create transaction', loading: false });
      throw error;
    }
  },

  updateTransaction: async (id: string, updates: Partial<Transaction>) => {
    set({ loading: true, error: null });
    try {
      const updatedTransaction = await transactionsApi.update(id, updates);
      set((state) => ({
        transactions: state.transactions.map((t) => 
          t.id === id ? updatedTransaction : t
        ),
        selectedTransaction: state.selectedTransaction?.id === id ? updatedTransaction : state.selectedTransaction,
        loading: false,
      }));
      return updatedTransaction;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update transaction', loading: false });
      throw error;
    }
  },

  deleteTransaction: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await transactionsApi.delete(id);
      set((state) => ({
        transactions: state.transactions.filter((t) => t.id !== id),
        selectedTransaction: state.selectedTransaction?.id === id ? null : state.selectedTransaction,
        loading: false,
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete transaction', loading: false });
      throw error;
    }
  },

  fetchSummary: async (filters?: TransactionFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const summary = await transactionsApi.getSummary(actualFilters);
      set({ summary, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch summary', loading: false });
      throw error;
    }
  },

  fetchAnomalyStats: async () => {
    set({ loading: true, error: null });
    try {
      const anomalyStats = await transactionsApi.getAnomalyStats();
      set({ anomalyStats, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch anomaly stats', loading: false });
      throw error;
    }
  },

  exportToCSV: async (filters?: TransactionFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const blob = await transactionsApi.exportToCSV(actualFilters);
      set({ loading: false });
      return blob;
    } catch (error: any) {
      set({ error: error.message || 'Failed to export transactions', loading: false });
      throw error;
    }
  },

  importFromCSV: async (file: File) => {
    set({ loading: true, error: null });
    try {
      const result = await transactionsApi.importFromCSV(file);
      // Refresh transactions after import
      await get().fetchTransactions();
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to import transactions', loading: false });
      throw error;
    }
  },
}));