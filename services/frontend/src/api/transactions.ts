import api from './client';

export interface Transaction {
  id: string;
  user_id: string;
  circle_id?: string;
  amount: number;
  currency: string;
  base_currency?: string;
  converted_amount?: number;
  description: string;
  private_description?: string;
  category: string;
  payment_method: string;
  transaction_date: string;
  is_private: boolean;
  hidden_until?: string;
  source: string;
  source_id?: string;
  is_anomaly: boolean;
  anomaly_type?: string;
  anomaly_severity?: number;
  matched_transaction_id?: string;
  notification_id?: string;
  parsed_from_text?: string;
  confidence_score?: number;
  created_at: string;
  updated_at: string;
}

export interface TransactionFilters {
  start_date?: string;
  end_date?: string;
  category?: string;
  payment_method?: string;
  is_private?: boolean;
  is_anomaly?: boolean;
  anomaly_type?: string;
  min_amount?: number;
  max_amount?: number;
  circle_id?: string;
}

export interface TransactionSummary {
  total_amount: number;
  transaction_count: number;
  average_amount: number;
  by_category: Record<string, number>;
  by_payment_method: Record<string, number>;
  by_date: Record<string, number>;
}

export interface TransactionAnomalyStats {
  total_anomalies: number;
  by_type: Record<string, number>;
  by_severity: Record<string, number>;
  resolved_count: number;
  pending_count: number;
}

export const transactionApi = {
  // Create a new transaction
  create: async (transaction: Omit<Transaction, 'id' | 'created_at' | 'updated_at'>) => {
    const response = await api.post<Transaction>('/api/transactions', transaction);
    return response.data;
  },

  // Get transaction by ID
  getById: async (id: string) => {
    const response = await api.get<Transaction>(`/api/transactions/${id}`);
    return response.data;
  },

  // Get all transactions with filters
  getAll: async (filters?: TransactionFilters) => {
    const response = await api.get<Transaction[]>('/api/transactions', { params: filters });
    return response.data;
  },

  // Update transaction
  update: async (id: string, updates: Partial<Transaction>) => {
    const response = await api.put<Transaction>(`/api/transactions/${id}`, updates);
    return response.data;
  },

  // Delete transaction
  delete: async (id: string) => {
    const response = await api.delete(`/api/transactions/${id}`);
    return response.data;
  },

  // Get transaction summary
  getSummary: async (filters?: TransactionFilters) => {
    const response = await api.get<TransactionSummary>('/api/transactions/summary', { params: filters });
    return response.data;
  },

  // Get monthly summary
  getMonthlySummary: async (year: number, month: number) => {
    const response = await api.get<TransactionSummary>(`/api/transactions/summary/monthly/${year}/${month}`);
    return response.data;
  },

  // Get category summary
  getCategorySummary: async (filters?: TransactionFilters) => {
    const response = await api.get<Record<string, number>>('/api/transactions/summary/category', { params: filters });
    return response.data;
  },

  // Get spending trends
  getSpendingTrends: async (period: 'day' | 'week' | 'month' | 'year', limit: number = 30) => {
    const response = await api.get<Record<string, number>>(`/api/transactions/trends/${period}?limit=${limit}`);
    return response.data;
  },

  // Get budget status
  getBudgetStatus: async (budget_id: string) => {
    const response = await api.get<{ spent: number; remaining: number; percentage: number }>(`/api/transactions/budget/${budget_id}/status`);
    return response.data;
  },

  // Find similar transactions
  findSimilar: async (transaction: Partial<Transaction>, threshold: number = 0.8) => {
    const response = await api.post<Transaction[]>('/api/transactions/similar', { transaction, threshold });
    return response.data;
  },

  // Match statement transaction
  matchStatementTransaction: async (statement_transaction: any) => {
    const response = await api.post<Transaction | null>('/api/transactions/match-statement', statement_transaction);
    return response.data;
  },

  // Create batch transactions
  createBatch: async (transactions: Omit<Transaction, 'id' | 'created_at' | 'updated_at'>[]) => {
    const response = await api.post<Transaction[]>('/api/transactions/batch', transactions);
    return response.data;
  },

  // Update batch transactions
  updateBatch: async (updates: Array<{ id: string; updates: Partial<Transaction> }>) => {
    const response = await api.put<Transaction[]>('/api/transactions/batch', updates);
    return response.data;
  },

  // Get uncategorized transactions
  getUncategorized: async () => {
    const response = await api.get<Transaction[]>('/api/transactions/uncategorized');
    return response.data;
  },

  // Get unverified transactions
  getUnverified: async () => {
    const response = await api.get<Transaction[]>('/api/transactions/unverified');
    return response.data;
  },

  // Soft delete transaction
  softDelete: async (id: string) => {
    const response = await api.delete(`/api/transactions/${id}/soft`);
    return response.data;
  },

  // Get transactions by family/circle
  getByFamily: async (circle_id: string, filters?: TransactionFilters) => {
    const response = await api.get<Transaction[]>(`/api/circles/${circle_id}/transactions`, { params: filters });
    return response.data;
  },

  // Get anomaly statistics
  getAnomalyStats: async () => {
    const response = await api.get<TransactionAnomalyStats>('/api/transactions/anomalies/stats');
    return response.data;
  },

  // Mark anomaly as resolved
  markAnomalyResolved: async (transaction_id: string, resolution_note?: string) => {
    const response = await api.post(`/api/transactions/${transaction_id}/resolve-anomaly`, { resolution_note });
    return response.data;
  },

  // Get duplicate detection results
  getDuplicates: async (threshold_days: number = 7, threshold_amount_percent: number = 5) => {
    const response = await api.get<Array<{ transactions: Transaction[]; confidence: number }>>(
      `/api/transactions/duplicates?threshold_days=${threshold_days}&threshold_amount_percent=${threshold_amount_percent}`
    );
    return response.data;
  },

  // Export transactions to CSV
  exportToCSV: async (filters?: TransactionFilters) => {
    const response = await api.get('/api/transactions/export/csv', { 
      params: filters,
      responseType: 'blob'
    });
    return response.data;
  },

  // Import transactions from CSV
  importFromCSV: async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post<{ imported: number; skipped: number; errors: string[] }>('/api/transactions/import/csv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },
};