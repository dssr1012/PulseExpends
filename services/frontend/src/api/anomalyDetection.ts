import api from './client';

export interface AnomalyDetectionRule {
  id: string;
  rule_name: string;
  rule_type: 'duplicate' | 'amount_mismatch' | 'orphan' | 'category' | 'time' | 'location' | 'custom';
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  threshold: number;
  parameters: Record<string, any>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DetectedAnomaly {
  id: string;
  rule_id: string;
  transaction_id: string;
  anomaly_type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  details: Record<string, any>;
  status: 'pending' | 'reviewed' | 'resolved' | 'ignored';
  resolved_by?: string;
  resolved_at?: string;
  resolution_note?: string;
  created_at: string;
  updated_at: string;
}

export interface AnomalyStats {
  total_anomalies: number;
  by_type: Record<string, number>;
  by_severity: Record<string, number>;
  by_status: Record<string, number>;
  pending_count: number;
  resolved_count: number;
  ignored_count: number;
  average_resolution_time_hours?: number;
}

export interface RuleEffectiveness {
  rule_id: string;
  rule_name: string;
  total_detected: number;
  true_positives: number;
  false_positives: number;
  effectiveness_rate: number;
  last_triggered: string;
}

export interface AnomalyFilters {
  rule_id?: string;
  transaction_id?: string;
  anomaly_type?: string;
  severity?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  resolved_by?: string;
}

export const anomalyDetectionApi = {
  // Rule Management
  createRule: async (rule: Omit<AnomalyDetectionRule, 'id' | 'created_at' | 'updated_at'>) => {
    const response = await api.post<AnomalyDetectionRule>('/api/anomaly-detection/rules', rule);
    return response.data;
  },

  getRuleById: async (id: string) => {
    const response = await api.get<AnomalyDetectionRule>(`/api/anomaly-detection/rules/${id}`);
    return response.data;
  },

  getAllRules: async () => {
    const response = await api.get<AnomalyDetectionRule[]>('/api/anomaly-detection/rules');
    return response.data;
  },

  updateRule: async (id: string, updates: Partial<AnomalyDetectionRule>) => {
    const response = await api.put<AnomalyDetectionRule>(`/api/anomaly-detection/rules/${id}`, updates);
    return response.data;
  },

  deleteRule: async (id: string) => {
    const response = await api.delete(`/api/anomaly-detection/rules/${id}`);
    return response.data;
  },

  activateRule: async (id: string) => {
    const response = await api.patch<AnomalyDetectionRule>(`/api/anomaly-detection/rules/${id}/activate`);
    return response.data;
  },

  deactivateRule: async (id: string) => {
    const response = await api.patch<AnomalyDetectionRule>(`/api/anomaly-detection/rules/${id}/deactivate`);
    return response.data;
  },

  // Anomaly Management
  createAnomaly: async (anomaly: Omit<DetectedAnomaly, 'id' | 'created_at' | 'updated_at'>) => {
    const response = await api.post<DetectedAnomaly>('/api/anomaly-detection/anomalies', anomaly);
    return response.data;
  },

  getAnomalyById: async (id: string) => {
    const response = await api.get<DetectedAnomaly>(`/api/anomaly-detection/anomalies/${id}`);
    return response.data;
  },

  getAnomaliesByCircle: async (circle_id: string, filters?: AnomalyFilters) => {
    const response = await api.get<DetectedAnomaly[]>(`/api/anomaly-detection/anomalies/circle/${circle_id}`, { params: filters });
    return response.data;
  },

  getAnomaliesByTransaction: async (transaction_id: string) => {
    const response = await api.get<DetectedAnomaly[]>(`/api/anomaly-detection/anomalies/transaction/${transaction_id}`);
    return response.data;
  },

  getAnomaliesByRule: async (rule_id: string, filters?: AnomalyFilters) => {
    const response = await api.get<DetectedAnomaly[]>(`/api/anomaly-detection/anomalies/rule/${rule_id}`, { params: filters });
    return response.data;
  },

  getAllAnomalies: async (filters?: AnomalyFilters) => {
    const response = await api.get<DetectedAnomaly[]>('/api/anomaly-detection/anomalies', { params: filters });
    return response.data;
  },

  updateAnomalyStatus: async (id: string, status: DetectedAnomaly['status'], resolution_note?: string) => {
    const response = await api.patch<DetectedAnomaly>(`/api/anomaly-detection/anomalies/${id}/status`, { status, resolution_note });
    return response.data;
  },

  deleteAnomaly: async (id: string) => {
    const response = await api.delete(`/api/anomaly-detection/anomalies/${id}`);
    return response.data;
  },

  // Detection and Analysis
  detectAnomalies: async (transaction_data: any) => {
    const response = await api.post<DetectedAnomaly[]>('/api/anomaly-detection/detect', transaction_data);
    return response.data;
  },

  batchDetectAnomalies: async (transactions: any[]) => {
    const response = await api.post<{ anomalies: DetectedAnomaly[]; processed: number; skipped: number }>(
      '/api/anomaly-detection/detect/batch',
      transactions
    );
    return response.data;
  },

  getAnomalyStats: async () => {
    const response = await api.get<AnomalyStats>('/api/anomaly-detection/stats');
    return response.data;
  },

  getRuleEffectiveness: async () => {
    const response = await api.get<RuleEffectiveness[]>('/api/anomaly-detection/rules/effectiveness');
    return response.data;
  },

  // Specific Anomaly Types
  detectDuplicates: async (transaction_data: any, threshold_days: number = 7, threshold_amount_percent: number = 5) => {
    const response = await api.post<Array<{ transaction: any; duplicate_of: any; confidence: number }>>('/api/anomaly-detection/detect/duplicates', {
      transaction: transaction_data,
      threshold_days,
      threshold_amount_percent,
    });
    return response.data;
  },

  detectAmountMismatch: async (transaction_data: any, expected_amount?: number, threshold_percent: number = 10) => {
    const response = await api.post<{ is_anomaly: boolean; actual_amount: number; expected_amount?: number; difference: number; difference_percent: number }>(
      '/api/anomaly-detection/detect/amount-mismatch',
      {
        transaction: transaction_data,
        expected_amount,
        threshold_percent,
      }
    );
    return response.data;
  },

  detectOrphanTransactions: async (circle_id: string, time_window_hours: number = 24) => {
    const response = await api.get<any[]>(`/api/anomaly-detection/detect/orphans/${circle_id}?time_window_hours=${time_window_hours}`);
    return response.data;
  },

  // Resolution Management
  bulkResolveAnomalies: async (anomaly_ids: string[], resolution_note?: string) => {
    const response = await api.post<{ resolved: number; failed: number }>('/api/anomaly-detection/anomalies/bulk-resolve', {
      anomaly_ids,
      resolution_note,
    });
    return response.data;
  },

  bulkIgnoreAnomalies: async (anomaly_ids: string[], reason?: string) => {
    const response = await api.post<{ ignored: number; failed: number }>('/api/anomaly-detection/anomalies/bulk-ignore', {
      anomaly_ids,
      reason,
    });
    return response.data;
  },

  // Export and Reports
  exportAnomaliesToCSV: async (filters?: AnomalyFilters) => {
    const response = await api.get('/api/anomaly-detection/anomalies/export/csv', {
      params: filters,
      responseType: 'blob',
    });
    return response.data;
  },

  generateAnomalyReport: async (start_date: string, end_date: string) => {
    const response = await api.get<{
      period: { start: string; end: string };
      stats: AnomalyStats;
      top_rules: RuleEffectiveness[];
      recent_anomalies: DetectedAnomaly[];
    }>(`/api/anomaly-detection/report?start_date=${start_date}&end_date=${end_date}`);
    return response.data;
  },
};