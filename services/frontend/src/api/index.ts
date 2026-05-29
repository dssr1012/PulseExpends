// Export all API services with proper namespacing
export { default as api } from './client';

// Auth API
export * as authApi from './auth';

// Circles API  
export * as circlesApi from './circles';

// Transactions API
export * as transactionsApi from './transactions';

// Credit Cards API
export * as creditCardsApi from './creditCards';

// Exchange Rates API
export * as exchangeRatesApi from './exchangeRates';

// Mobile Notifications API
export * as mobileNotificationsApi from './mobileNotifications';

// Anomaly Detection API
export * as anomalyDetectionApi from './anomalyDetection';

// Notification Whitelist API
export * as notificationWhitelistApi from './notificationWhitelist';