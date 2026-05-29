import React, { useEffect, useState } from 'react';
import { useTransactionStore } from '../store/transactionStore';
import { useCreditCardStore } from '../store/creditCardStore';
import { useExchangeRateStore } from '../store/exchangeRateStore';
import { useMobileNotificationStore } from '../store/mobileNotificationStore';
import { useAnomalyDetectionStore } from '../store/anomalyDetectionStore';
import { useNotificationWhitelistStore } from '../store/notificationWhitelistStore';
import { 
  BarChart3, CreditCard, DollarSign, Bell, AlertTriangle, Shield,
  TrendingUp, TrendingDown, Users, FileText, Download, Upload,
  Calendar, Filter, Search, Plus, MoreVertical
} from 'lucide-react';

const Dashboard = () => {
  const transactionStore = useTransactionStore();
  const creditCardStore = useCreditCardStore();
  const exchangeRateStore = useExchangeRateStore();
  const notificationStore = useMobileNotificationStore();
  const anomalyStore = useAnomalyDetectionStore();
  const whitelistStore = useNotificationWhitelistStore();

  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');

  useEffect(() => {
    const fetchDashboardData = async () => {
      setLoading(true);
      try {
        await Promise.all([
          transactionStore.fetchSummary(),
          transactionStore.fetchAnomalyStats(),
          creditCardStore.fetchCreditCards(),
          creditCardStore.fetchUpcomingPayments(),
          exchangeRateStore.fetchLatestRates(),
          exchangeRateStore.fetchSupportedCurrencies(),
          notificationStore.fetchStats(),
          notificationStore.fetchPendingNotifications(),
          anomalyStore.fetchAnomalyStats(),
          anomalyStore.fetchAllRules(),
          whitelistStore.fetchStats(),
        ]);
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 p-6 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading dashboard data...</p>
        </div>
      </div>
    );
  }

  const stats = [
    {
      title: 'Total Transactions',
      value: transactionStore.summary?.transaction_count || 0,
      change: '+12.5%',
      icon: BarChart3,
      color: 'bg-blue-500',
      trend: 'up' as const,
    },
    {
      title: 'Credit Cards',
      value: creditCardStore.creditCards.length,
      change: '+2',
      icon: CreditCard,
      color: 'bg-green-500',
      trend: 'up' as const,
    },
    {
      title: 'Exchange Rates',
      value: exchangeRateStore.latestRates.length,
      change: 'Live',
      icon: DollarSign,
      color: 'bg-purple-500',
      trend: 'neutral' as const,
    },
    {
      title: 'Pending Notifications',
      value: notificationStore.notifications.filter(n => n.status === 'pending').length,
      change: '-5',
      icon: Bell,
      color: 'bg-yellow-500',
      trend: 'down' as const,
    },
    {
      title: 'Active Anomalies',
      value: anomalyStore.anomalies.filter(a => a.status === 'pending').length,
      change: '+3',
      icon: AlertTriangle,
      color: 'bg-red-500',
      trend: 'up' as const,
    },
    {
      title: 'Whitelisted Apps',
      value: whitelistStore.whitelistEntries.filter(e => e.is_whitelisted).length,
      change: '+1',
      icon: Shield,
      color: 'bg-indigo-500',
      trend: 'up' as const,
    },
  ];

  const recentTransactions = transactionStore.transactions.slice(0, 5);
  const upcomingPayments = creditCardStore.upcomingPayments.slice(0, 3);
  const highConfidenceNotifications = notificationStore.notifications
    .filter(n => n.confidence_score > 0.8)
    .slice(0, 3);

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">PulseExpends Dashboard</h1>
        <p className="text-gray-600 mt-2">Monitor and manage your financial ecosystem</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-6 mb-8">
        {stats.map((stat, index) => (
          <div key={index} className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center justify-between">
              <div className={`p-3 rounded-lg ${stat.color} bg-opacity-10`}>
                <stat.icon className={`h-6 w-6 ${stat.color.replace('bg-', 'text-')}`} />
              </div>
              <div className="flex items-center space-x-1">
                {stat.trend === 'up' && <TrendingUp className="h-4 w-4 text-green-500" />}
                {stat.trend === 'down' && <TrendingDown className="h-4 w-4 text-red-500" />}
                <span className={`text-sm font-medium ${
                  stat.trend === 'up' ? 'text-green-600' : 
                  stat.trend === 'down' ? 'text-red-600' : 'text-gray-600'
                }`}>
                  {stat.change}
                </span>
              </div>
            </div>
            <h3 className="text-2xl font-bold text-gray-900 mt-4">{stat.value.toLocaleString()}</h3>
            <p className="text-gray-600 text-sm mt-1">{stat.title}</p>
          </div>
        ))}
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column - Transactions & Payments */}
        <div className="lg:col-span-2 space-y-8">
          {/* Recent Transactions */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Recent Transactions</h2>
                <p className="text-gray-600 text-sm">Latest financial activities</p>
              </div>
              <button className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                <Plus className="h-4 w-4" />
                <span>New Transaction</span>
              </button>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-3 px-4 text-gray-600 font-medium">Description</th>
                    <th className="text-left py-3 px-4 text-gray-600 font-medium">Amount</th>
                    <th className="text-left py-3 px-4 text-gray-600 font-medium">Category</th>
                    <th className="text-left py-3 px-4 text-gray-600 font-medium">Date</th>
                    <th className="text-left py-3 px-4 text-gray-600 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {recentTransactions.map((transaction) => (
                    <tr key={transaction.id} className="border-b border-gray-100 hover:bg-gray-50">
                      <td className="py-3 px-4">
                        <div className="font-medium text-gray-900">{transaction.description}</div>
                        {transaction.is_private && (
                          <span className="text-xs text-gray-500">Private</span>
                        )}
                      </td>
                      <td className="py-3 px-4">
                        <div className={`font-medium ${transaction.amount < 0 ? 'text-red-600' : 'text-green-600'}`}>
                          {transaction.currency} {Math.abs(transaction.amount).toFixed(2)}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-1 text-xs rounded-full bg-blue-100 text-blue-800">
                          {transaction.category}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-gray-600">
                        {new Date(transaction.transaction_date).toLocaleDateString()}
                      </td>
                      <td className="py-3 px-4">
                        {transaction.is_anomaly ? (
                          <span className="px-2 py-1 text-xs rounded-full bg-red-100 text-red-800">Anomaly</span>
                        ) : (
                          <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Normal</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            
            {recentTransactions.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                No transactions found
              </div>
            )}
          </div>

          {/* Upcoming Payments */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Upcoming Payments</h2>
                <p className="text-gray-600 text-sm">Credit card payments due soon</p>
              </div>
              <button className="flex items-center space-x-2 px-4 py-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors">
                <Calendar className="h-4 w-4" />
                <span>View All</span>
              </button>
            </div>
            
            <div className="space-y-4">
              {upcomingPayments.map((payment, index) => (
                <div key={index} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center space-x-4">
                    <div className="p-2 bg-blue-100 rounded-lg">
                      <CreditCard className="h-5 w-5 text-blue-600" />
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">Credit Card Payment</div>
                      <div className="text-sm text-gray-600">Due {new Date(payment.due_date).toLocaleDateString()}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-gray-900">${payment.minimum_payment.toFixed(2)}</div>
                    <div className="text-sm text-gray-600">Minimum payment</div>
                  </div>
                </div>
              ))}
              
              {upcomingPayments.length === 0 && (
                <div className="text-center py-4 text-gray-500">
                  No upcoming payments
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Column - Notifications & Quick Actions */}
        <div className="space-y-8">
          {/* High Confidence Notifications */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">High Confidence Notifications</h2>
                <p className="text-gray-600 text-sm">Ready for processing</p>
              </div>
              <Bell className="h-5 w-5 text-yellow-500" />
            </div>
            
            <div className="space-y-4">
              {highConfidenceNotifications.map((notification) => (
                <div key={notification.id} className="p-4 border border-gray-200 rounded-lg">
                  <div className="flex justify-between items-start">
                    <div>
                      <div className="font-medium text-gray-900">{notification.app_name}</div>
                      <div className="text-sm text-gray-600 mt-1">{notification.notification_text}</div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className={`px-2 py-1 text-xs rounded-full ${
                        notification.confidence_score > 0.9 ? 'bg-green-100 text-green-800' :
                        notification.confidence_score > 0.7 ? 'bg-yellow-100 text-yellow-800' :
                        'bg-red-100 text-red-800'
                      }`}>
                        {(notification.confidence_score * 100).toFixed(0)}%
                      </div>
                      <button className="p-1 hover:bg-gray-100 rounded">
                        <MoreVertical className="h-4 w-4 text-gray-500" />
                      </button>
                    </div>
                  </div>
                  {notification.parsed_amount && (
                    <div className="mt-2 text-sm">
                      <span className="text-gray-600">Amount: </span>
                      <span className="font-medium text-green-600">
                        {notification.parsed_currency} {notification.parsed_amount.toFixed(2)}
                      </span>
                    </div>
                  )}
                </div>
              ))}
              
              {highConfidenceNotifications.length === 0 && (
                <div className="text-center py-4 text-gray-500">
                  No high confidence notifications
                </div>
              )}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-6">Quick Actions</h2>
            
            <div className="grid grid-cols-2 gap-4">
              <button className="flex flex-col items-center justify-center p-4 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors">
                <div className="p-3 bg-blue-100 rounded-lg mb-2">
                  <Plus className="h-6 w-6 text-blue-600" />
                </div>
                <span className="font-medium text-gray-900">Add Transaction</span>
                <span className="text-sm text-gray-600 mt-1">Manual entry</span>
              </button>
              
              <button className="flex flex-col items-center justify-center p-4 bg-green-50 hover:bg-green-100 rounded-lg transition-colors">
                <div className="p-3 bg-green-100 rounded-lg mb-2">
                  <Upload className="h-6 w-6 text-green-600" />
                </div>
                <span className="font-medium text-gray-900">Import CSV</span>
                <span className="text-sm text-gray-600 mt-1">Batch upload</span>
              </button>
              
              <button className="flex flex-col items-center justify-center p-4 bg-purple-50 hover:bg-purple-100 rounded-lg transition-colors">
                <div className="p-3 bg-purple-100 rounded-lg mb-2">
                  <DollarSign className="h-6 w-6 text-purple-600" />
                </div>
                <span className="font-medium text-gray-900">Convert Currency</span>
                <span className="text-sm text-gray-600 mt-1">Real-time rates</span>
              </button>
              
              <button className="flex flex-col items-center justify-center p-4 bg-red-50 hover:bg-red-100 rounded-lg transition-colors">
                <div className="p-3 bg-red-100 rounded-lg mb-2">
                  <AlertTriangle className="h-6 w-6 text-red-600" />
                </div>
                <span className="font-medium text-gray-900">Check Anomalies</span>
                <span className="text-sm text-gray-600 mt-1">Review alerts</span>
              </button>
            </div>
          </div>

          {/* System Status */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">System Status</h2>
            
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-gray-700">API Service</span>
                <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Online</span>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-gray-700">Database</span>
                <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Connected</span>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-gray-700">Exchange Rates</span>
                <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Live</span>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-gray-700">Notification Parser</span>
                <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Active</span>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="text-gray-700">Anomaly Detection</span>
                <span className="px-2 py-1 text-xs rounded-full bg-yellow-100 text-yellow-800">Monitoring</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;