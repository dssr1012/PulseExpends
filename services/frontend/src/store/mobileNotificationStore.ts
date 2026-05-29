import { create } from 'zustand';
import { MobileNotification, NotificationFilters, NotificationStats, AppStats } from '../api/mobileNotifications';
import { mobileNotificationsApi } from '../api';

interface MobileNotificationState {
  notifications: MobileNotification[];
  selectedNotification: MobileNotification | null;
  filters: NotificationFilters;
  stats: NotificationStats | null;
  appStats: AppStats[];
  loading: boolean;
  error: string | null;
  
  // Actions
  setNotifications: (notifications: MobileNotification[]) => void;
  setSelectedNotification: (notification: MobileNotification | null) => void;
  setFilters: (filters: NotificationFilters) => void;
  setStats: (stats: NotificationStats | null) => void;
  setAppStats: (appStats: AppStats[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // API Actions
  fetchNotifications: (filters?: NotificationFilters) => Promise<void>;
  fetchNotificationById: (id: string) => Promise<void>;
  fetchNotificationsByUser: (userId: string, filters?: NotificationFilters) => Promise<void>;
  fetchNotificationsByDevice: (deviceId: string, filters?: NotificationFilters) => Promise<void>;
  fetchPendingNotifications: () => Promise<void>;
  createNotification: (notification: Omit<MobileNotification, 'id' | 'created_at' | 'updated_at'>) => Promise<MobileNotification>;
  updateNotification: (id: string, updates: Partial<MobileNotification>) => Promise<MobileNotification>;
  deleteNotification: (id: string) => Promise<void>;
  updateNotificationStatus: (id: string, status: MobileNotification['status']) => Promise<MobileNotification>;
  updateConfidenceScore: (id: string, confidenceScore: number) => Promise<MobileNotification>;
  markAsProcessed: (id: string, transactionId?: string) => Promise<MobileNotification>;
  fetchStats: () => Promise<void>;
  fetchAppStats: (appPackage?: string) => Promise<void>;
  parseNotification: (notificationText: string, appPackage: string) => Promise<{
    amount?: number;
    currency?: string;
    merchant?: string;
    category?: string;
    confidence: number;
    parsed_text: string;
  }>;
  processBatch: (notificationIds: string[], status: MobileNotification['status'], transactionId?: string) => Promise<{ processed: number; failed: number }>;
  fetchHighConfidenceNotifications: (threshold?: number) => Promise<void>;
  linkToTransaction: (notificationId: string, transactionId: string) => Promise<MobileNotification>;
  unlinkTransaction: (notificationId: string) => Promise<MobileNotification>;
  exportToCSV: (filters?: NotificationFilters) => Promise<Blob>;
  importFromSync: (notifications: Omit<MobileNotification, 'id' | 'created_at' | 'updated_at'>[]) => Promise<{ imported: number; duplicates: number; errors: string[] }>;
  cleanupOldNotifications: (daysOld?: number) => Promise<{ deleted: number }>;
}

export const useMobileNotificationStore = create<MobileNotificationState>((set, get) => ({
  notifications: [],
  selectedNotification: null,
  filters: {},
  stats: null,
  appStats: [],
  loading: false,
  error: null,

  setNotifications: (notifications) => set({ notifications }),
  setSelectedNotification: (selectedNotification) => set({ selectedNotification }),
  setFilters: (filters) => set({ filters }),
  setStats: (stats) => set({ stats }),
  setAppStats: (appStats) => set({ appStats }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  fetchNotifications: async (filters?: NotificationFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const notifications = await mobileNotificationsApi.getAll(actualFilters);
      set({ notifications, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch notifications', loading: false });
      throw error;
    }
  },

  fetchNotificationById: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const notification = await mobileNotificationsApi.getById(id);
      set({ selectedNotification: notification, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch notification', loading: false });
      throw error;
    }
  },

  fetchNotificationsByUser: async (userId: string, filters?: NotificationFilters) => {
    set({ loading: true, error: null });
    try {
      const notifications = await mobileNotificationsApi.getByUser(userId, filters);
      set({ notifications, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch user notifications', loading: false });
      throw error;
    }
  },

  fetchNotificationsByDevice: async (deviceId: string, filters?: NotificationFilters) => {
    set({ loading: true, error: null });
    try {
      const notifications = await mobileNotificationsApi.getByDevice(deviceId, filters);
      set({ notifications, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch device notifications', loading: false });
      throw error;
    }
  },

  fetchPendingNotifications: async () => {
    set({ loading: true, error: null });
    try {
      const notifications = await mobileNotificationsApi.getPending();
      set({ notifications, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch pending notifications', loading: false });
      throw error;
    }
  },

  createNotification: async (notificationData) => {
    set({ loading: true, error: null });
    try {
      const notification = await mobileNotificationsApi.create(notificationData);
      set((state) => ({ 
        notifications: [notification, ...state.notifications],
        loading: false 
      }));
      return notification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create notification', loading: false });
      throw error;
    }
  },

  updateNotification: async (id: string, updates: Partial<MobileNotification>) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.update(id, updates);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === id ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === id ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update notification', loading: false });
      throw error;
    }
  },

  deleteNotification: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await mobileNotificationsApi.delete(id);
      set((state) => ({
        notifications: state.notifications.filter((n) => n.id !== id),
        selectedNotification: state.selectedNotification?.id === id ? null : state.selectedNotification,
        loading: false,
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete notification', loading: false });
      throw error;
    }
  },

  updateNotificationStatus: async (id: string, status: MobileNotification['status']) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.updateStatus(id, status);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === id ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === id ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update notification status', loading: false });
      throw error;
    }
  },

  updateConfidenceScore: async (id: string, confidenceScore: number) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.updateConfidence(id, confidenceScore);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === id ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === id ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update confidence score', loading: false });
      throw error;
    }
  },

  markAsProcessed: async (id: string, transactionId?: string) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.markAsProcessed(id, transactionId);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === id ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === id ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to mark notification as processed', loading: false });
      throw error;
    }
  },

  fetchStats: async () => {
    set({ loading: true, error: null });
    try {
      const stats = await mobileNotificationsApi.getStats();
      set({ stats, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch notification stats', loading: false });
      throw error;
    }
  },

  fetchAppStats: async (appPackage?: string) => {
    set({ loading: true, error: null });
    try {
      const appStats = await mobileNotificationsApi.getAppStats(appPackage);
      set({ appStats, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch app stats', loading: false });
      throw error;
    }
  },

  parseNotification: async (notificationText: string, appPackage: string) => {
    set({ loading: true, error: null });
    try {
      const result = await mobileNotificationsApi.parseNotification(notificationText, appPackage);
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to parse notification', loading: false });
      throw error;
    }
  },

  processBatch: async (notificationIds: string[], status: MobileNotification['status'], transactionId?: string) => {
    set({ loading: true, error: null });
    try {
      const result = await mobileNotificationsApi.processBatch(notificationIds, status, transactionId);
      // Refresh notifications after batch processing
      await get().fetchNotifications();
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to process batch', loading: false });
      throw error;
    }
  },

  fetchHighConfidenceNotifications: async (threshold: number = 0.8) => {
    set({ loading: true, error: null });
    try {
      const notifications = await mobileNotificationsApi.getHighConfidence(threshold);
      set({ notifications, loading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to fetch high confidence notifications', loading: false });
      throw error;
    }
  },

  linkToTransaction: async (notificationId: string, transactionId: string) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.linkToTransaction(notificationId, transactionId);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === notificationId ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === notificationId ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to link notification to transaction', loading: false });
      throw error;
    }
  },

  unlinkTransaction: async (notificationId: string) => {
    set({ loading: true, error: null });
    try {
      const updatedNotification = await mobileNotificationsApi.unlinkTransaction(notificationId);
      set((state) => ({
        notifications: state.notifications.map((n) => 
          n.id === notificationId ? updatedNotification : n
        ),
        selectedNotification: state.selectedNotification?.id === notificationId ? updatedNotification : state.selectedNotification,
        loading: false,
      }));
      return updatedNotification;
    } catch (error: any) {
      set({ error: error.message || 'Failed to unlink transaction', loading: false });
      throw error;
    }
  },

  exportToCSV: async (filters?: NotificationFilters) => {
    set({ loading: true, error: null });
    try {
      const actualFilters = filters || get().filters;
      const blob = await mobileNotificationsApi.exportToCSV(actualFilters);
      set({ loading: false });
      return blob;
    } catch (error: any) {
      set({ error: error.message || 'Failed to export notifications', loading: false });
      throw error;
    }
  },

  importFromSync: async (notifications) => {
    set({ loading: true, error: null });
    try {
      const result = await mobileNotificationsApi.importFromSync(notifications);
      // Refresh notifications after import
      await get().fetchNotifications();
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to import notifications', loading: false });
      throw error;
    }
  },

  cleanupOldNotifications: async (daysOld: number = 30) => {
    set({ loading: true, error: null });
    try {
      const result = await mobileNotificationsApi.cleanupOld(daysOld);
      // Refresh notifications after cleanup
      await get().fetchNotifications();
      set({ loading: false });
      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to cleanup old notifications', loading: false });
      throw error;
    }
  },
}));