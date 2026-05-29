import api from './client';

export interface MobileNotification {
  id: string;
  user_id: string;
  device_id?: string;
  app_name: string;
  app_package: string;
  notification_title: string;
  notification_text: string;
  notification_time: string;
  parsed_amount?: number;
  parsed_currency?: string;
  parsed_merchant?: string;
  parsed_category?: string;
  confidence_score: number;
  status: 'pending' | 'processed' | 'ignored' | 'error';
  transaction_id?: string;
  raw_data?: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationStats {
  total_notifications: number;
  by_app: Record<string, number>;
  by_status: Record<string, number>;
  by_day: Record<string, number>;
  average_confidence: number;
  processed_count: number;
  ignored_count: number;
  error_count: number;
}

export interface AppStats {
  app_name: string;
  notification_count: number;
  average_confidence: number;
  processed_rate: number;
  last_notification: string;
}

export interface NotificationFilters {
  app_name?: string;
  app_package?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  min_confidence?: number;
  max_confidence?: number;
  has_transaction?: boolean;
}

export const mobileNotificationApi = {
  // Create a new mobile notification
  create: async (notification: Omit<MobileNotification, 'id' | 'created_at' | 'updated_at'>) => {
    const response = await api.post<MobileNotification>('/api/mobile-notifications', notification);
    return response.data;
  },

  // Get notification by ID
  getById: async (id: string) => {
    const response = await api.get<MobileNotification>(`/api/mobile-notifications/${id}`);
    return response.data;
  },

  // Get all notifications with filters
  getAll: async (filters?: NotificationFilters) => {
    const response = await api.get<MobileNotification[]>('/api/mobile-notifications', { params: filters });
    return response.data;
  },

  // Update notification
  update: async (id: string, updates: Partial<MobileNotification>) => {
    const response = await api.put<MobileNotification>(`/api/mobile-notifications/${id}`, updates);
    return response.data;
  },

  // Delete notification
  delete: async (id: string) => {
    const response = await api.delete(`/api/mobile-notifications/${id}`);
    return response.data;
  },

  // Get notifications by user
  getByUser: async (user_id: string, filters?: NotificationFilters) => {
    const response = await api.get<MobileNotification[]>(`/api/mobile-notifications/user/${user_id}`, { params: filters });
    return response.data;
  },

  // Get notifications by device
  getByDevice: async (device_id: string, filters?: NotificationFilters) => {
    const response = await api.get<MobileNotification[]>(`/api/mobile-notifications/device/${device_id}`, { params: filters });
    return response.data;
  },

  // Get pending notifications
  getPending: async () => {
    const response = await api.get<MobileNotification[]>('/api/mobile-notifications/pending');
    return response.data;
  },

  // Update notification status
  updateStatus: async (id: string, status: MobileNotification['status']) => {
    const response = await api.patch<MobileNotification>(`/api/mobile-notifications/${id}/status`, { status });
    return response.data;
  },

  // Update confidence score
  updateConfidence: async (id: string, confidence_score: number) => {
    const response = await api.patch<MobileNotification>(`/api/mobile-notifications/${id}/confidence`, { confidence_score });
    return response.data;
  },

  // Mark notification as processed
  markAsProcessed: async (id: string, transaction_id?: string) => {
    const response = await api.post<MobileNotification>(`/api/mobile-notifications/${id}/process`, { transaction_id });
    return response.data;
  },

  // Get notification statistics
  getStats: async () => {
    const response = await api.get<NotificationStats>('/api/mobile-notifications/stats');
    return response.data;
  },

  // Get app-specific statistics
  getAppStats: async (app_package?: string) => {
    const response = await api.get<AppStats[]>(`/api/mobile-notifications/stats/apps${app_package ? `?app_package=${app_package}` : ''}`);
    return response.data;
  },

  // Parse notification text (extract amount, currency, merchant, category)
  parseNotification: async (notification_text: string, app_package: string) => {
    const response = await api.post<{
      amount?: number;
      currency?: string;
      merchant?: string;
      category?: string;
      confidence: number;
      parsed_text: string;
    }>('/api/mobile-notifications/parse', { notification_text, app_package });
    return response.data;
  },

  // Bulk process notifications
  processBatch: async (notification_ids: string[], status: MobileNotification['status'], transaction_id?: string) => {
    const response = await api.post<{ processed: number; failed: number }>('/api/mobile-notifications/process-batch', {
      notification_ids,
      status,
      transaction_id,
    });
    return response.data;
  },

  // Get notifications with high confidence (for auto-processing)
  getHighConfidence: async (threshold: number = 0.8) => {
    const response = await api.get<MobileNotification[]>(`/api/mobile-notifications/high-confidence?threshold=${threshold}`);
    return response.data;
  },

  // Get notifications by confidence range
  getByConfidenceRange: async (min: number, max: number) => {
    const response = await api.get<MobileNotification[]>(`/api/mobile-notifications/confidence-range?min=${min}&max=${max}`);
    return response.data;
  },

  // Link notification to transaction
  linkToTransaction: async (notification_id: string, transaction_id: string) => {
    const response = await api.post<MobileNotification>(`/api/mobile-notifications/${notification_id}/link-transaction`, { transaction_id });
    return response.data;
  },

  // Unlink notification from transaction
  unlinkTransaction: async (notification_id: string) => {
    const response = await api.patch<MobileNotification>(`/api/mobile-notifications/${notification_id}/unlink-transaction`);
    return response.data;
  },

  // Get notifications by transaction
  getByTransaction: async (transaction_id: string) => {
    const response = await api.get<MobileNotification[]>(`/api/mobile-notifications/transaction/${transaction_id}`);
    return response.data;
  },

  // Export notifications to CSV
  exportToCSV: async (filters?: NotificationFilters) => {
    const response = await api.get('/api/mobile-notifications/export/csv', {
      params: filters,
      responseType: 'blob',
    });
    return response.data;
  },

  // Import notifications from device sync
  importFromSync: async (notifications: Omit<MobileNotification, 'id' | 'created_at' | 'updated_at'>[]) => {
    const response = await api.post<{ imported: number; duplicates: number; errors: string[] }>(
      '/api/mobile-notifications/import/sync',
      notifications
    );
    return response.data;
  },

  // Clean up old notifications
  cleanupOld: async (days_old: number = 30) => {
    const response = await api.delete<{ deleted: number }>(`/api/mobile-notifications/cleanup?days_old=${days_old}`);
    return response.data;
  },
};