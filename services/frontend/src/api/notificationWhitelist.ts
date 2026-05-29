import api from './client';

export interface NotificationAppWhitelist {
  id: string;
  user_id: string;
  circle_id?: string;
  app_name: string;
  app_package: string;
  platform: 'android' | 'ios' | 'web';
  is_whitelisted: boolean;
  auto_whitelist: boolean;
  confidence_threshold: number;
  last_used?: string;
  usage_count: number;
  created_at: string;
  updated_at: string;
}

export interface WhitelistStats {
  total_apps: number;
  whitelisted_count: number;
  auto_whitelisted_count: number;
  by_platform: Record<string, number>;
  top_apps: Array<{ app_name: string; usage_count: number; last_used: string }>;
}

export interface AppValidationResult {
  app_name: string;
  app_package: string;
  platform: 'android' | 'ios' | 'web';
  is_valid: boolean;
  validation_errors?: string[];
  suggested_app_name?: string;
  confidence_score: number;
}

export interface WhitelistFilters {
  user_id?: string;
  circle_id?: string;
  platform?: string;
  is_whitelisted?: boolean;
  auto_whitelist?: boolean;
  app_name?: string;
  app_package?: string;
}

export const notificationWhitelistApi = {
  // Create a new app whitelist entry
  create: async (entry: Omit<NotificationAppWhitelist, 'id' | 'created_at' | 'updated_at' | 'usage_count'>) => {
    const response = await api.post<NotificationAppWhitelist>('/api/notification-whitelist', entry);
    return response.data;
  },

  // Get whitelist entry by ID
  getById: async (id: string) => {
    const response = await api.get<NotificationAppWhitelist>(`/api/notification-whitelist/${id}`);
    return response.data;
  },

  // Get all whitelist entries with filters
  getAll: async (filters?: WhitelistFilters) => {
    const response = await api.get<NotificationAppWhitelist[]>('/api/notification-whitelist', { params: filters });
    return response.data;
  },

  // Update whitelist entry
  update: async (id: string, updates: Partial<NotificationAppWhitelist>) => {
    const response = await api.put<NotificationAppWhitelist>(`/api/notification-whitelist/${id}`, updates);
    return response.data;
  },

  // Delete whitelist entry
  delete: async (id: string) => {
    const response = await api.delete(`/api/notification-whitelist/${id}`);
    return response.data;
  },

  // Get whitelist entries by user
  getByUser: async (user_id: string, filters?: WhitelistFilters) => {
    const response = await api.get<NotificationAppWhitelist[]>(`/api/notification-whitelist/user/${user_id}`, { params: filters });
    return response.data;
  },

  // Get whitelist entries by circle
  getByCircle: async (circle_id: string, filters?: WhitelistFilters) => {
    const response = await api.get<NotificationAppWhitelist[]>(`/api/notification-whitelist/circle/${circle_id}`, { params: filters });
    return response.data;
  },

  // Check if app is whitelisted
  isWhitelisted: async (app_package: string, user_id?: string, circle_id?: string) => {
    const params: any = { app_package };
    if (user_id) params.user_id = user_id;
    if (circle_id) params.circle_id = circle_id;
    
    const response = await api.get<{ is_whitelisted: boolean; entry?: NotificationAppWhitelist }>('/api/notification-whitelist/check', { params });
    return response.data;
  },

  // Whitelist an app
  whitelistApp: async (app_package: string, app_name: string, platform: 'android' | 'ios' | 'web', user_id?: string, circle_id?: string) => {
    const response = await api.post<NotificationAppWhitelist>('/api/notification-whitelist/whitelist', {
      app_package,
      app_name,
      platform,
      user_id,
      circle_id,
    });
    return response.data;
  },

  // Remove app from whitelist
  removeFromWhitelist: async (app_package: string, user_id?: string, circle_id?: string) => {
    const response = await api.delete<{ removed: boolean; entry?: NotificationAppWhitelist }>('/api/notification-whitelist/remove', {
      data: { app_package, user_id, circle_id },
    });
    return response.data;
  },

  // Toggle auto-whitelist
  toggleAutoWhitelist: async (id: string, auto_whitelist: boolean) => {
    const response = await api.patch<NotificationAppWhitelist>(`/api/notification-whitelist/${id}/auto-whitelist`, { auto_whitelist });
    return response.data;
  },

  // Update confidence threshold
  updateConfidenceThreshold: async (id: string, confidence_threshold: number) => {
    const response = await api.patch<NotificationAppWhitelist>(`/api/notification-whitelist/${id}/confidence-threshold`, { confidence_threshold });
    return response.data;
  },

  // Get whitelist statistics
  getStats: async () => {
    const response = await api.get<WhitelistStats>('/api/notification-whitelist/stats');
    return response.data;
  },

  // Get app usage statistics
  getAppUsageStats: async (app_package?: string) => {
    const response = await api.get<Array<{ app_name: string; app_package: string; usage_count: number; last_used: string; is_whitelisted: boolean }>>(
      `/api/notification-whitelist/stats/usage${app_package ? `?app_package=${app_package}` : ''}`
    );
    return response.data;
  },

  // Validate app package (Android package name validation)
  validateAppPackage: async (app_package: string, platform: 'android' | 'ios' | 'web') => {
    const response = await api.post<AppValidationResult>('/api/notification-whitelist/validate', { app_package, platform });
    return response.data;
  },

  // Auto-whitelist based on usage
  autoWhitelistByUsage: async (min_usage_count: number = 5, min_confidence: number = 0.7) => {
    const response = await api.post<{ whitelisted: number; skipped: number; errors: string[] }>('/api/notification-whitelist/auto-whitelist', {
      min_usage_count,
      min_confidence,
    });
    return response.data;
  },

  // Bulk update whitelist status
  bulkUpdate: async (updates: Array<{ app_package: string; is_whitelisted: boolean; user_id?: string; circle_id?: string }>) => {
    const response = await api.put<{ updated: number; failed: number }>('/api/notification-whitelist/bulk', updates);
    return response.data;
  },

  // Import whitelist from CSV
  importFromCSV: async (file: File, platform: 'android' | 'ios' | 'web') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('platform', platform);
    
    const response = await api.post<{ imported: number; duplicates: number; errors: string[] }>('/api/notification-whitelist/import/csv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  // Export whitelist to CSV
  exportToCSV: async (filters?: WhitelistFilters) => {
    const response = await api.get('/api/notification-whitelist/export/csv', {
      params: filters,
      responseType: 'blob',
    });
    return response.data;
  },

  // Get suggested app names (for auto-completion)
  getSuggestedAppNames: async (partial_name: string, limit: number = 10) => {
    const response = await api.get<Array<{ app_name: string; app_package: string; platform: string; usage_count: number }>>(
      `/api/notification-whitelist/suggestions?partial_name=${partial_name}&limit=${limit}`
    );
    return response.data;
  },

  // Clean up unused whitelist entries
  cleanupUnused: async (days_unused: number = 90) => {
    const response = await api.delete<{ removed: number }>(`/api/notification-whitelist/cleanup?days_unused=${days_unused}`);
    return response.data;
  },

  // Reset usage count for an app
  resetUsageCount: async (id: string) => {
    const response = await api.patch<NotificationAppWhitelist>(`/api/notification-whitelist/${id}/reset-usage`);
    return response.data;
  },

  // Increment usage count (called when app sends notification)
  incrementUsage: async (app_package: string, user_id?: string, circle_id?: string) => {
    const response = await api.post<NotificationAppWhitelist>('/api/notification-whitelist/increment-usage', {
      app_package,
      user_id,
      circle_id,
    });
    return response.data;
  },

  // Get platform-specific whitelist (Android vs iOS vs Web)
  getByPlatform: async (platform: 'android' | 'ios' | 'web', filters?: WhitelistFilters) => {
    const response = await api.get<NotificationAppWhitelist[]>(`/api/notification-whitelist/platform/${platform}`, { params: filters });
    return response.data;
  },
};