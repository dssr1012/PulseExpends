// PulseExpends Authentication System
// Handles user authentication, registration, and session management

const AUTH_API_URL = window.location.hostname === 'pulseexpends.duckdns.org' ? 
    'http://api.pulseexpends.duckdns.org/api' : 
    'http://localhost:8082/api';

class AuthService {
    constructor() {
        this.token = localStorage.getItem('auth_token');
        this.user = JSON.parse(localStorage.getItem('auth_user') || 'null');
        this.isAuthenticated = !!this.token;
    }

    // Check if user is authenticated
    isLoggedIn() {
        return this.isAuthenticated;
    }

    // Get current user
    getCurrentUser() {
        return this.user;
    }

    // Get auth token
    getToken() {
        return this.token;
    }

    // Login with email/password
    async login(email, password) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Login failed');
            }

            const data = await response.json();
            
            // Save token and user data
            this.token = data.token;
            this.user = data.user;
            this.isAuthenticated = true;
            
            localStorage.setItem('auth_token', data.token);
            localStorage.setItem('auth_user', JSON.stringify(data.user));
            
            return { success: true, data };
        } catch (error) {
            console.error('Login error:', error);
            return { success: false, error: error.message };
        }
    }

    // Register new user
    async register(email, password, username, fullName) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password, username, fullName })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Registration failed');
            }

            const data = await response.json();
            
            // Save token and user data
            this.token = data.token;
            this.user = data.user;
            this.isAuthenticated = true;
            
            localStorage.setItem('auth_token', data.token);
            localStorage.setItem('auth_user', JSON.stringify(data.user));
            
            return { success: true, data };
        } catch (error) {
            console.error('Registration error:', error);
            return { success: false, error: error.message };
        }
    }

    // Login with Google
    async loginWithGoogle() {
        window.location.href = `${AUTH_API_URL}/auth/google`;
    }

    // Logout
    async logout() {
        try {
            if (this.token) {
                await fetch(`${AUTH_API_URL}/auth/logout`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.token}`,
                        'Content-Type': 'application/json'
                    }
                });
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            // Clear local storage
            localStorage.removeItem('auth_token');
            localStorage.removeItem('auth_user');
            
            // Reset state
            this.token = null;
            this.user = null;
            this.isAuthenticated = false;
            
            // Redirect to login
            window.location.href = '/auth/login.html';
        }
    }

    // Get user profile
    async getProfile() {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/auth/profile`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch profile');
            }

            const data = await response.json();
            this.user = data.user;
            localStorage.setItem('auth_user', JSON.stringify(data.user));
            
            return { success: true, data };
        } catch (error) {
            console.error('Profile fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Update user profile
    async updateProfile(updates) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/auth/profile`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(updates)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Update failed');
            }

            const data = await response.json();
            this.user = data.user;
            localStorage.setItem('auth_user', JSON.stringify(data.user));
            
            return { success: true, data };
        } catch (error) {
            console.error('Profile update error:', error);
            return { success: false, error: error.message };
        }
    }

    // Change password
    async changePassword(currentPassword, newPassword) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/auth/change-password`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ currentPassword, newPassword })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Password change failed');
            }

            return { success: true };
        } catch (error) {
            console.error('Password change error:', error);
            return { success: false, error: error.message };
        }
    }

    // Verify token validity
    async verifyToken() {
        if (!this.token) {
            return { success: false, valid: false };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/auth/verify`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            return { success: true, valid: response.ok };
        } catch (error) {
            console.error('Token verification error:', error);
            return { success: false, valid: false };
        }
    }

    // Get circles for current user
    async getCircles() {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch circles');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Circles fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Create new circle
    async createCircle(name, description, currency = 'USD', isPublic = false) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name, description, currency, isPublic })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to create circle');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Circle creation error:', error);
            return { success: false, error: error.message };
        }
    }

    // Join circle with invitation code
    async joinCircle(code) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles/join/${code}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to join circle');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Circle join error:', error);
            return { success: false, error: error.message };
        }
    }

    // Get circle members
    async getCircleMembers(circleId) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles/${circleId}/members`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch circle members');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Circle members fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Add member to circle
    async addCircleMember(circleId, userId, role = 'member') {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles/${circleId}/members`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ userId, role })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to add member');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Add member error:', error);
            return { success: false, error: error.message };
        }
    }

    // Invite to circle by email
    async inviteToCircle(circleId, email, role = 'member') {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/circles/${circleId}/invite`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, role })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to send invitation');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Invite error:', error);
            return { success: false, error: error.message };
        }
    }

    // Get transactions
    async getTransactions(filters = {}) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const queryParams = new URLSearchParams(filters).toString();
            const url = `${AUTH_API_URL}/transactions${queryParams ? `?${queryParams}` : ''}`;
            
            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch transactions');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Transactions fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Create transaction
    async createTransaction(transactionData) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/transactions`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(transactionData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to create transaction');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Transaction creation error:', error);
            return { success: false, error: error.message };
        }
    }

    // Get user stats
    async getUserStats() {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/stats`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch stats');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Stats fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Get circle stats
    async getCircleStats(circleId) {
        if (!this.token) {
            return { success: false, error: 'Not authenticated' };
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/stats/circle/${circleId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    return { success: false, error: 'Session expired' };
                }
                throw new Error('Failed to fetch circle stats');
            }

            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('Circle stats fetch error:', error);
            return { success: false, error: error.message };
        }
    }

    // Check if we're in a callback from Google OAuth
    checkOAuthCallback() {
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');
        
        if (token) {
            // Save token and redirect to dashboard
            this.token = token;
            this.isAuthenticated = true;
            
            // Get user profile
            this.getProfile().then(result => {
                if (result.success) {
                    localStorage.setItem('auth_token', token);
                    localStorage.setItem('auth_user', JSON.stringify(result.data.user));
                    window.location.href = '/dashboard.html';
                } else {
                    console.error('Failed to get profile after OAuth:', result.error);
                    window.location.href = '/auth/login.html';
                }
            });
            
            return true;
        }
        
        return false;
    }
}

// Create global auth instance
const auth = new AuthService();

// Export for use in other files
if (typeof module !== 'undefined' && module.exports) {
    module.exports = auth;
} else {
    window.auth = auth;
}

// Initialize auth on page load
document.addEventListener('DOMContentLoaded', function() {
    // Check for OAuth callback
    if (auth.checkOAuthCallback()) {
        return;
    }
    
    // Check if user is authenticated and redirect if needed
    const currentPath = window.location.pathname;
    const isAuthPage = currentPath.includes('/auth/');
    const isLoginPage = currentPath.endsWith('/auth/login.html');
    const isRegisterPage = currentPath.endsWith('/auth/register.html');
    
    if (auth.isLoggedIn() && (isLoginPage || isRegisterPage)) {
        // Redirect to dashboard if already logged in
        window.location.href = '/dashboard.html';
    } else if (!auth.isLoggedIn() && !isAuthPage && !currentPath.endsWith('/index.html')) {
        // Redirect to login if not authenticated
        window.location.href = '/auth/login.html';
    }
    
    // Update UI based on auth state
    updateAuthUI();
});

// Update UI based on authentication state
function updateAuthUI() {
    const user = auth.getCurrentUser();
    
    // Update user info in navbar if elements exist
    const userAvatar = document.getElementById('userAvatar');
    const userName = document.getElementById('userName');
    const userEmail = document.getElementById('userEmail');
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    const userMenu = document.getElementById('userMenu');
    
    if (auth.isLoggedIn() && user) {
        // User is logged in
        if (userAvatar) {
            if (user.avatarURL) {
                userAvatar.src = user.avatarURL;
                userAvatar.style.display = 'block';
            } else {
                userAvatar.style.display = 'none';
            }
        }
        
        if (userName) {
            userName.textContent = user.fullName || user.username || user.email;
        }
        
        if (userEmail) {
            userEmail.textContent = user.email;
        }
        
        if (loginBtn) loginBtn.style.display = 'none';
        if (logoutBtn) logoutBtn.style.display = 'block';
        if (userMenu) userMenu.style.display = 'flex';
    } else {
        // User is not logged in
        if (loginBtn) loginBtn.style.display = 'block';
        if (logoutBtn) logoutBtn.style.display = 'none';
        if (userMenu) userMenu.style.display = 'none';
    }
}

// Logout handler
function handleLogout() {
    auth.logout();
}

// Make functions available globally
window.handleLogout = handleLogout;
window.auth = auth;