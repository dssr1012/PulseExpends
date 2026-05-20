// PulseExpends Authentication Module
// Handles user authentication, sessions, and protected routes

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
    async checkAuth() {
        if (!this.token) {
            return false;
        }

        try {
            const response = await fetch(`${AUTH_API_URL}/auth/profile`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.user = data.user;
                localStorage.setItem('auth_user', JSON.stringify(data.user));
                this.isAuthenticated = true;
                return true;
            } else {
                // Token expired or invalid
                this.logout();
                return false;
            }
        } catch (error) {
            console.error('Auth check failed:', error);
            this.logout();
            return false;
        }
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

            const data = await response.json();

            if (response.ok) {
                this.token = data.token;
                this.user = data.user;
                this.isAuthenticated = true;
                
                localStorage.setItem('auth_token', data.token);
                localStorage.setItem('auth_user', JSON.stringify(data.user));
                
                return { success: true, user: data.user };
            } else {
                return { success: false, error: data.message || 'Error al iniciar sesión' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Register new user
    async register(userData) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userData)
            });

            const data = await response.json();

            if (response.ok) {
                this.token = data.token;
                this.user = data.user;
                this.isAuthenticated = true;
                
                localStorage.setItem('auth_token', data.token);
                localStorage.setItem('auth_user', JSON.stringify(data.user));
                
                return { success: true, user: data.user };
            } else {
                return { success: false, error: data.message || 'Error al registrar usuario' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Login with Google
    loginWithGoogle() {
        window.location.href = `${AUTH_API_URL}/auth/google`;
    }

    // Handle Google callback
    async handleGoogleCallback(token) {
        try {
            // Verify token with backend
            const response = await fetch(`${AUTH_API_URL}/auth/verify-google`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token })
            });

            const data = await response.json();

            if (response.ok) {
                this.token = data.token;
                this.user = data.user;
                this.isAuthenticated = true;
                
                localStorage.setItem('auth_token', data.token);
                localStorage.setItem('auth_user', JSON.stringify(data.user));
                
                return { success: true, user: data.user };
            } else {
                return { success: false, error: data.message || 'Error al verificar token de Google' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Logout
    logout() {
        this.token = null;
        this.user = null;
        this.isAuthenticated = false;
        
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
        
        // Call logout endpoint
        fetch(`${AUTH_API_URL}/auth/logout`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        }).catch(console.error);
        
        // Redirect to login
        window.location.href = '/auth/login.html';
    }

    // Get current user
    getCurrentUser() {
        return this.user;
    }

    // Get auth token
    getToken() {
        return this.token;
    }

    // Check if user is authenticated
    isLoggedIn() {
        return this.isAuthenticated;
    }

    // Update user profile
    async updateProfile(profileData) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/profile`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(profileData)
            });

            const data = await response.json();

            if (response.ok) {
                this.user = { ...this.user, ...data.user };
                localStorage.setItem('auth_user', JSON.stringify(this.user));
                return { success: true, user: data.user };
            } else {
                return { success: false, error: data.message || 'Error al actualizar perfil' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Change password
    async changePassword(currentPassword, newPassword) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/change-password`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ currentPassword, newPassword })
            });

            const data = await response.json();

            if (response.ok) {
                return { success: true, message: data.message };
            } else {
                return { success: false, error: data.message || 'Error al cambiar contraseña' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Forgot password
    async forgotPassword(email) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/forgot-password`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email })
            });

            const data = await response.json();

            if (response.ok) {
                return { success: true, message: data.message };
            } else {
                return { success: false, error: data.message || 'Error al solicitar recuperación' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Reset password
    async resetPassword(token, newPassword) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/reset-password`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token, newPassword })
            });

            const data = await response.json();

            if (response.ok) {
                return { success: true, message: data.message };
            } else {
                return { success: false, error: data.message || 'Error al restablecer contraseña' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Get user sessions
    async getSessions() {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/sessions`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            const data = await response.json();

            if (response.ok) {
                return { success: true, sessions: data.sessions };
            } else {
                return { success: false, error: data.message || 'Error al obtener sesiones' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Revoke session
    async revokeSession(sessionId) {
        try {
            const response = await fetch(`${AUTH_API_URL}/auth/sessions/${sessionId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            const data = await response.json();

            if (response.ok) {
                return { success: true, message: data.message };
            } else {
                return { success: false, error: data.message || 'Error al revocar sesión' };
            }
        } catch (error) {
            return { success: false, error: 'Error de conexión. Por favor, intenta nuevamente.' };
        }
    }

    // Make authenticated API requests
    async fetchWithAuth(url, options = {}) {
        const headers = {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json',
            ...options.headers
        };

        const response = await fetch(url, {
            ...options,
            headers
        });

        // Handle 401 Unauthorized
        if (response.status === 401) {
            this.logout();
            throw new Error('Sesión expirada. Por favor, inicia sesión nuevamente.');
        }

        return response;
    }
}

// Create global auth instance
window.auth = new AuthService();

// Protected route middleware
function requireAuth() {
    if (!window.auth.isLoggedIn()) {
        // Store current URL to redirect back after login
        const currentPath = window.location.pathname + window.location.search;
        if (currentPath !== '/auth/login.html' && currentPath !== '/auth/register.html') {
            sessionStorage.setItem('redirectAfterLogin', currentPath);
        }
        window.location.href = '/auth/login.html';
        return false;
    }
    return true;
}

// Check Google callback on page load
document.addEventListener('DOMContentLoaded', function() {
    // Handle Google callback
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    
    if (token && window.location.pathname.includes('callback')) {
        window.auth.handleGoogleCallback(token).then(result => {
            if (result.success) {
                // Redirect to dashboard or stored URL
                const redirectUrl = sessionStorage.getItem('redirectAfterLogin') || '/dashboard.html';
                sessionStorage.removeItem('redirectAfterLogin');
                window.location.href = redirectUrl;
            } else {
                alert('Error al iniciar sesión con Google: ' + result.error);
                window.location.href = '/auth/login.html';
            }
        });
    }
    
    // Check auth on protected pages
    const protectedPages = ['dashboard.html', 'transactions.html', 'circles.html', 'upload.html', 'settings.html', 'profile.html'];
    const currentPage = window.location.pathname.split('/').pop();
    
    if (protectedPages.includes(currentPage)) {
        window.auth.checkAuth().then(isAuthenticated => {
            if (!isAuthenticated) {
                requireAuth();
            } else {
                // Update UI with user info
                const user = window.auth.getCurrentUser();
                if (user) {
                    // Update user name in navbar if element exists
                    const userNameElement = document.getElementById('userName');
                    if (userNameElement && user.fullName) {
                        userNameElement.textContent = user.fullName.split(' ')[0];
                    }
                    
                    // Update user avatar if element exists
                    const userAvatarElement = document.getElementById('userAvatar');
                    if (userAvatarElement) {
                        if (user.avatarURL) {
                            userAvatarElement.style.backgroundImage = `url('${user.avatarURL}')`;
                            userAvatarElement.innerHTML = '';
                        } else {
                            userAvatarElement.textContent = user.fullName ? user.fullName.charAt(0).toUpperCase() : 'U';
                        }
                    }
                }
            }
        });
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { AuthService, requireAuth };
}