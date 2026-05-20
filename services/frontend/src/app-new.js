// PulseExpends Frontend Application
// Main JavaScript file for the Fintonic-like expense management system

// Import auth service
// Note: Since we're using ES modules, we need to load auth.js as a module
// For now, we'll include the auth service inline and refactor later

// API Configuration with subdomains
// In production, use subdomains
// In development, fallback to relative paths
const isProduction = window.location.hostname === 'pulseexpends.duckdns.org' || 
                     window.location.hostname === 'www.pulseexpends.duckdns.org';

const API_BASE_URL = isProduction ? 'http://api.pulseexpends.duckdns.org' : '/api/mcp';
const PDF_API_URL = isProduction ? 'http://pdf.pulseexpends.duckdns.org' : '/api/pdf';
const AUTH_API_URL = isProduction ? 'http://api.pulseexpends.duckdns.org/api' : 'http://localhost:8082/api';

// Global state
let transactions = [];
let summary = {};
let currentPage = 'dashboard';
let currentCircle = null;
let userCircles = [];

// DOM Elements
const mainContent = document.getElementById('main-content');
const loadingElement = document.getElementById('loading');
const addExpenseModal = document.getElementById('addExpenseModal');
const uploadModal = document.getElementById('uploadModal');
const expenseForm = document.getElementById('expenseForm');
const pdfFileInput = document.getElementById('pdfFile');
const dropArea = document.getElementById('dropArea');
const uploadStatus = document.getElementById('uploadStatus');

// Auth service (simplified version for now)
class AuthService {
    constructor() {
        this.token = localStorage.getItem('auth_token');
        this.user = JSON.parse(localStorage.getItem('auth_user') || 'null');
        this.isAuthenticated = !!this.token;
    }

    isLoggedIn() {
        return this.isAuthenticated;
    }

    getCurrentUser() {
        return this.user;
    }

    getToken() {
        return this.token;
    }

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
            localStorage.removeItem('auth_token');
            localStorage.removeItem('auth_user');
            this.token = null;
            this.user = null;
            this.isAuthenticated = false;
            window.location.href = '/auth/login.html';
        }
    }

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
}

const auth = new AuthService();

// Initialize the application
async function initApp() {
    // Check authentication
    if (!auth.isLoggedIn()) {
        window.location.href = '/auth/login.html';
        return;
    }

    // Set today's date as default
    const dateInput = document.getElementById('date');
    if (dateInput) {
        dateInput.valueAsDate = new Date();
    }
    
    // Setup navigation
    setupNavigation();
    
    // Setup modal handlers
    setupModals();
    
    // Setup expense form
    setupExpenseForm();
    
    // Setup PDF upload
    setupPDFUpload();
    
    // Load user circles
    await loadUserCircles();
    
    // Update user info in navbar
    updateAuthUI();
    
    // Load initial page
    loadPage('dashboard');
    
    // Check API connectivity
    checkAPIConnectivity();
}

// Setup navigation
function setupNavigation() {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            link.classList.add('active');
            currentPage = link.dataset.page;
            loadPage(currentPage);
        });
    });
}

// Setup modals
function setupModals() {
    const closeModalBtn = document.getElementById('closeModal');
    const closeUploadModalBtn = document.getElementById('closeUploadModal');
    
    if (closeModalBtn) {
        closeModalBtn.addEventListener('click', () => {
            addExpenseModal.classList.remove('active');
        });
    }
    
    if (closeUploadModalBtn) {
        closeUploadModalBtn.addEventListener('click', () => {
            uploadModal.classList.remove('active');
        });
    }

    window.addEventListener('click', (e) => {
        if (e.target === addExpenseModal) addExpenseModal.classList.remove('active');
        if (e.target === uploadModal) uploadModal.classList.remove('active');
    });
}

// Setup expense form
function setupExpenseForm() {
    if (!expenseForm) return;
    
    expenseForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const circleSelect = document.getElementById('circle');
        const circleId = circleSelect ? circleSelect.value : null;
        
        const expenseData = {
            description: document.getElementById('description').value,
            amount: parseFloat(document.getElementById('amount').value),
            category: document.getElementById('category').value,
            date: document.getElementById('date').value + 'T00:00:00Z',
            type: document.getElementById('type').value,
            currency: 'USD',
            circleId: circleId || null
        };

        try {
            const result = await auth.createTransaction(expenseData);
            
            if (result.success) {
                alert('¡Gasto agregado exitosamente!');
                addExpenseModal.classList.remove('active');
                expenseForm.reset();
                document.getElementById('date').valueAsDate = new Date();
                loadDashboard();
            } else {
                throw new Error(result.error || 'Error al agregar gasto');
            }
        } catch (error) {
            alert('Error: ' + error.message);
        }
    });
}

// Setup PDF upload
function setupPDFUpload() {
    if (!pdfFileInput || !dropArea) return;
    
    pdfFileInput.addEventListener('change', handleFileUpload);
    
    // Drag and drop for PDF
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false);
    });

    dropArea.addEventListener('drop', handleDrop, false);
}

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

function highlight() {
    dropArea.style.borderColor = '#00a8ff';
    dropArea.style.backgroundColor = '#f8f9ff';
}

function unhighlight() {
    dropArea.style.borderColor = '#ddd';
    dropArea.style.backgroundColor = 'transparent';
}

function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;
    handleFiles(files);
}

function handleFileUpload(e) {
    const files = e.target.files;
    handleFiles(files);
}

async function handleFiles(files) {
    if (files.length === 0) return;
    
    const file = files[0];
    if (file.type !== 'application/pdf') {
        alert('Por favor, sube solo archivos PDF');
        return;
    }

    if (file.size > 10 * 1024 * 1024) { // 10MB limit
        alert('El archivo es demasiado grande (máximo 10MB)');
        return;
    }

    uploadStatus.style.display = 'block';
    
    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`${PDF_API_URL}/parse`, {
            method: 'POST',
            body: formData
        });

        const result = await response.json();
        
        if (response.ok) {
            uploadStatus.innerHTML = `
                <div style="background: #d4edda; color: #155724; padding: 15px; border-radius: 8px;">
                    <h4><i class="fas fa-check-circle"></i> PDF procesado exitosamente</h4>
                    <p>${result.message}</p>
                    <p><strong>Texto extraído:</strong> ${result.data?.text?.substring(0, 200) || 'No se extrajo texto'}...</p>
                </div>
            `;
            
            // Simulate extracting transactions from PDF
            setTimeout(() => {
                simulatePDFTransactions(result.data?.text || '');
                uploadStatus.style.display = 'none';
                uploadModal.classList.remove('active');
                loadDashboard();
            }, 2000);
        } else {
            throw new Error(result.error || 'Error al procesar el PDF');
        }
    } catch (error) {
        uploadStatus.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 15px; border-radius: 8px;">
                <h4><i class="fas fa-exclamation-circle"></i> Error al procesar PDF</h4>
                <p>${error.message}</p>
            </div>
        `;
    }
}

function simulatePDFTransactions(pdfText) {
    // In a real application, this would parse the PDF text and extract transactions
    // For now, we'll simulate adding some transactions
    const simulatedTransactions = [
        {
            description: "Supermercado (PDF)",
            amount: 85.50,
            category: "Food",
            type: "expense"
        },
        {
            description: "Gasolina (PDF)",
            amount: 45.00,
            category: "Transportation",
            type: "expense"
        },
        {
            description: "Restaurante (PDF)",
            amount: 65.75,
            category: "Food",
            type: "expense"
        }
    ];

    // Add simulated transactions
    simulatedTransactions.forEach(async (tx, index) => {
        setTimeout(async () => {
            await auth.createTransaction({
                ...tx,
                currency: 'USD',
                date: new Date().toISOString()
            });
        }, index * 500);
    });
}

// Page loading functions
async function loadPage(page) {
    if (loadingElement) loadingElement.style.display = 'flex';
    
    switch(page) {
        case 'dashboard':
            await loadDashboard();
            break;
        case 'transactions':
            await loadTransactions();
            break;
        case 'categories':
            await loadCategories();
            break;
        case 'upload':
            loadUpload();
            break;
        case 'alerts':
            loadAlerts();
            break;
        case 'settings':
            loadSettings();
            break;
        case 'circles':
            await loadCirclesPage();
            break;
    }
    
    if (loadingElement) loadingElement.style.display = 'none';
}

async function loadDashboard() {
    try {
        // Load transactions and summary
        const [transactionsResult, summaryResult] = await Promise.all([
            auth.getTransactions(),
            auth.getUserStats()
        ]);

        if (!transactionsResult.success || !summaryResult.success) {
            throw new Error('Error al cargar datos del servidor');
        }

        transactions = transactionsResult.data?.transactions || [];
        summary = summaryResult.data?.stats || {};

        // Calculate stats
        const totalExpenses = summary.totalExpenses || 0;
        const totalIncome = summary.totalIncome || 0;
        const netBalance = summary.balance || 0;
        const transactionCount = summary.transactionCount || 0;

        // Get recent transactions (last 5)
        const recentTransactions = transactions.slice(-5).reverse();

        // Render dashboard
        mainContent.innerHTML = `
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-header">
                        <h3 class="stat-title">Gastos Totales</h3>
                        <div class="stat-icon expense">
                            <i class="fas fa-money-bill-wave"></i>
                        </div>
                    </div>
                    <div class="stat-value">$${totalExpenses.toFixed(2)}</div>
                    <div class="stat-change negative">
                        <i class="fas fa-arrow-down"></i>
                        <span>Este mes</span>
                    </div>
                </div>
                
                <div class="stat-card">
                    <div class="stat-header">
                        <h3 class="stat-title">Ingresos Totales</h3>
                        <div class="stat-icon income">
                            <i class="fas fa-wallet"></i>
                        </div>
                    </div>
                    <div class="stat-value">$${totalIncome.toFixed(2)}</div>
                    <div class="stat-change positive">
                        <i class="fas fa-arrow-up"></i>
                        <span>Este mes</span>
                    </div>
                </div>
                
                <div class="stat-card">
                    <div class="stat-header">
                        <h3 class="stat-title">Balance Neto</h3>
                        <div class="stat-icon balance">
                            <i class="fas fa-balance-scale"></i>
                        </div>
                    </div>
                    <div class="stat-value">$${netBalance.toFixed(2)}</div>
                    <div class="stat-change ${netBalance >= 0 ? 'positive' : 'negative'}">
                        <i class="fas fa-${netBalance >= 0 ? 'arrow-up' : 'arrow-down'}"></i>
                        <span>${netBalance >= 0 ? 'Positivo' : 'Negativo'}</span>
                    </div>
                </div>
                
                <div class="stat-card">
                    <div class="stat-header">
                        <h3 class="stat-title">Transacciones</h3>
                        <div class="stat-icon transactions">
                            <i class="fas fa-receipt"></i>
                        </div>
                    </div>
                    <div class="stat-value">${transactionCount}</div>
                    <div class="stat-change positive">
                        <i class="fas fa-chart-line"></i>
                        <span>Total registradas</span>
                    </div>
                </div>
            </div>

            <div class="recent-transactions">
                <div class="section-header">
                    <h2 class="section-title">Transacciones Recientes</h2>
                    <button class="btn btn-primary" id="addExpenseBtn">
                        <i class="fas fa-plus"></i> Agregar Gasto
                    </button>
                </div>
                
                <div class="transactions-list" id="transactionsList">
                    ${recentTransactions.length > 0 ? 
                        recentTransactions.map(tx => `
                            <div class="transaction-item">
                                <div class="transaction-info">
                                    <div class="transaction-icon ${getCategoryClass(tx.category)}">
                                        <i class="${getCategoryIcon(tx.category)}"></i>
                                    </div>
                                    <div class="transaction-details">
                                        <h4>${tx.description || 'Sin descripción'}</h4>
                                        <p>${formatDate(tx.date)} • ${tx.category}</p>
                                    </div>
                                </div>
                                <div class="transaction-amount ${tx.type}">
                                    ${tx.type === 'expense' ? '-' : '+'}$${tx.amount.toFixed(2)}
                                </div>
                            </div>
                        `).join('') :
                        '<p style="text-align: center; color: var(--gray-color); padding: 40px;">No hay transacciones recientes</p>'
                    }
                </div>
            </div>

            <div class="categories-chart">
                <div class="section-header">
                    <h2 class="section-title">Gastos por Categoría</h2>
                    <button class="btn btn-success" onclick="loadPage('categories')">
                        <i class="fas fa-chart-pie"></i> Ver Detalles
                    </button>
                </div>
                
                <div class="chart-container" id="chartContainer">
                    <div style="display: flex; flex-direction: column; gap: 10px;">
                        ${Object.keys(summary).length > 0 ? `
                            <div class="category-item">
                                <div class="category-info">
                                    <div class="category-color" style="background-color: #FF6B6B"></div>
                                    <span class="category-name">Gastos</span>
                                </div>
                                <span class="category-percentage">$${totalExpenses.toFixed(2)}</span>
                            </div>
                            <div class="category-item">
                                <div class="category-info">
                                    <div class="category-color" style="background-color: #4ECDC4"></div>
                                    <span class="category-name">Ingresos</span>
                                </div>
                                <span class="category-percentage">$${totalIncome.toFixed(2)}</span>
                            </div>
                            <div class="category-item">
                                <div class="category-info">
                                    <div class="category-color" style="background-color: #FFD166"></div>
                                    <span class="category-name">Balance</span>
                                </div>
                                <span class="category-percentage">$${netBalance.toFixed(2)}</span>
                            </div>
                        ` : 
                        '<p style="text-align: center; color: var(--gray-color); padding: 20px;">No hay datos de categorías</p>'}
                    </div>
                </div>
            </div>

            <div class="upload-section">
                <div class="section-header">
                    <h2 class="section-title">Subir Resumen de Tarjeta</h2>
                    <button class="btn btn-primary" onclick="uploadModal.classList.add('active')">
                        <i class="fas fa-file-upload"></i> Subir PDF
                    </button>
                </div>
                <p style="color: var(--gray-color); margin-bottom: 15px;">
                    Sube el resumen de tu tarjeta de crédito en PDF para extraer automáticamente los gastos.
                </p>
                <div class="upload-area" onclick="uploadModal.classList.add('active')">
                    <div class="upload-icon">
                        <i class="fas fa-file-pdf"></i>
                    </div>
                    <h3>Arrastra tu archivo PDF aquí</h3>
                    <p class="upload-text">o haz clic para seleccionar</p>
                    <button class="upload-button">
                        Seleccionar Archivo
                    </button>
                </div>
            </div>

            <div class="alerts-section">
                <div class="section-header">
                    <h2 class="section-title">Alertas Recientes</h2>
                    <button class="btn" onclick="loadPage('alerts')" style="background: var(--warning-color); color: white;">
                        <i class="fas fa-bell"></i> Ver Todas
                    </button>
                </div>
                
                <div class="alert-item">
                    <div class="alert-icon">
                        <i class="fas fa-exclamation-triangle"></i>
                    </div>
                    <div class="alert-content">
                        <h4>Gasto alto en Comida</h4>
                        <p>Has gastado $150 en comida este mes, superando tu presupuesto de $100</p>
                    </div>
                </div>
                
                <div class="alert-item">
                    <div class="alert-icon">
                        <i class="fas fa-info-circle"></i>
                    </div>
                    <div class="alert-content">
                        <h4>Resumen pendiente</h4>
                        <p>Tu resumen de tarjeta de crédito de Mayo está listo para revisión</p>
                    </div>
                </div>
            </div>
        `;

        // Add event listener to the add expense button
        const addExpenseBtn = document.getElementById('addExpenseBtn');
        if (addExpenseBtn) {
            addExpenseBtn.addEventListener('click', () => {
                addExpenseModal.classList.add('active');
            });
        }

    } catch (error) {
        console.error('Error loading dashboard:', error);
        mainContent.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 20px; border-radius: 8px; text-align: center;">
                <h3><i class="fas fa-exclamation-circle"></i> Error al cargar datos</h3>
                <p>No se pudo conectar con el servidor. Por favor, intenta de nuevo más tarde.</p>
                <button class="btn btn-primary" onclick="loadDashboard()" style="margin-top: 15px;">
                    <i class="fas fa-redo"></i> Reintentar
                </button>
            </div>
        `;
    }
}

async function loadTransactions() {
    try {
        const result = await auth.getTransactions();
        if (!result.success) throw new Error(result.error || 'Error al cargar transacciones');
        
        transactions = result.data?.transactions || [];

        mainContent.innerHTML = `
            <div class="recent-transactions" style="grid-column: 1 / -1;">
                <div class="section-header">
                    <h2 class="section-title">Todas las Transacciones</h2>
                    <div>
                        <button class="btn btn-primary" id="addExpenseBtn">
                            <i class="fas fa-plus"></i> Agregar Gasto
                        </button>
                        <button class="btn" onclick="uploadModal.classList.add('active')" style="margin-left: 10px;">
                            <i class="fas fa-file-upload"></i> Subir PDF
                        </button>
                    </div>
                </div>
                
                <div style="margin-bottom: 20px; display: flex; gap: 10px;">
                    <input type="text" id="searchTransactions" class="form-input" placeholder="Buscar transacciones..." style="flex: 1;">
                    <select id="filterCategory" class="form-select" style="width: 200px;">
                        <option value="">Todas las categorías</option>
                        <option value="Food">Alimentos</option>
                        <option value="Transportation">Transporte</option>
                        <option value="Shopping">Compras</option>
                        <option value="Entertainment">Entretenimiento</option>
                        <option value="Coffee">Café</option>
                        <option value="Utilities">Servicios</option>
                        <option value="Healthcare">Salud</option>
                        <option value="Education">Educación</option>
                        <option value="Other">Otros</option>
                    </select>
                    <select id="filterType" class="form-select" style="width: 150px;">
                        <option value="">Todos los tipos</option>
                        <option value="expense">Gastos</option>
                        <option value="income">Ingresos</option>
                    </select>
                </div>
                
                <div class="transactions-list" id="allTransactionsList">
                    ${transactions.length > 0 ? 
                        transactions.reverse().map(tx => `
                            <div class="transaction-item">
                                <div class="transaction-info">
                                    <div class="transaction-icon ${getCategoryClass(tx.category)}">
                                        <i class="${getCategoryIcon(tx.category)}"></i>
                                    </div>
                                    <div class="transaction-details">
                                        <h4>${tx.description || 'Sin descripción'}</h4>
                                        <p>${formatDate(tx.date)} • ${tx.category} • ${tx.type === 'expense' ? 'Gasto' : 'Ingreso'}</p>
                                    </div>
                                </div>
                                <div class="transaction-amount ${tx.type}">
                                    ${tx.type === 'expense' ? '-' : '+'}$${tx.amount.toFixed(2)}
                                </div>
                            </div>
                        `).join('') :
                        '<p style="text-align: center; color: var(--gray-color); padding: 40px;">No hay transacciones registradas</p>'
                    }
                </div>
            </div>
        `;

        const addExpenseBtn = document.getElementById('addExpenseBtn');
        if (addExpenseBtn) {
            addExpenseBtn.addEventListener('click', () => {
                addExpenseModal.classList.add('active');
            });
        }

        // Add filtering functionality
        const searchInput = document.getElementById('searchTransactions');
        const filterCategory = document.getElementById('filterCategory');
        const filterType = document.getElementById('filterType');

        function filterTransactions() {
            const searchTerm = searchInput.value.toLowerCase();
            const categoryFilter = filterCategory.value;
            const typeFilter = filterType.value;

            const filtered = transactions.filter(tx => {
                const matchesSearch = (tx.description?.toLowerCase().includes(searchTerm) || 
                                      tx.category?.toLowerCase().includes(searchTerm)) || false;
                const matchesCategory = !categoryFilter || tx.category === categoryFilter;
                const matchesType = !typeFilter || tx.type === typeFilter;
                
                return matchesSearch && matchesCategory && matchesType;
            });

            const transactionsList = document.getElementById('allTransactionsList');
            transactionsList.innerHTML = filtered.length > 0 ? 
                filtered.reverse().map(tx => `
                    <div class="transaction-item">
                        <div class="transaction-info">
                            <div class="transaction-icon ${getCategoryClass(tx.category)}">
                                <i class="${getCategoryIcon(tx.category)}"></i>
                            </div>
                            <div class="transaction-details">
                                <h4>${tx.description || 'Sin descripción'}</h4>
                                <p>${formatDate(tx.date)} • ${tx.category} • ${tx.type === 'expense' ? 'Gasto' : 'Ingreso'}</p>
                            </div>
                        </div>
                        <div class="transaction-amount ${tx.type}">
                            ${tx.type === 'expense' ? '-' : '+'}$${tx.amount.toFixed(2)}
                        </div>
                    </div>
                `).join('') :
                '<p style="text-align: center; color: var(--gray-color); padding: 40px;">No se encontraron transacciones</p>';
        }

        if (searchInput) searchInput.addEventListener('input', filterTransactions);
        if (filterCategory) filterCategory.addEventListener('change', filterTransactions);
        if (filterType) filterType.addEventListener('change', filterTransactions);

    } catch (error) {
        console.error('Error loading transactions:', error);
        mainContent.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 20px; border-radius: 8px; text-align: center;">
                <h3><i class="fas fa-exclamation-circle"></i> Error al cargar transacciones</h3>
                <p>No se pudo conectar con el servidor.</p>
                <button class="btn btn-primary" onclick="loadTransactions()" style="margin-top: 15px;">
                    <i class="fas fa-redo"></i> Reintentar
                </button>
            </div>
        `;
    }
}

async function loadCategories() {
    try {
        const result = await auth.getUserStats();
        if (!result.success) throw new Error(result.error || 'Error al cargar categorías');
        
        const stats = result.data?.stats || {};
        const totalExpenses = stats.totalExpenses || 1;
        
        // For now, use dummy categories - in a real app, you'd get these from the API
        const categories = {
            'Food': totalExpenses * 0.4,
            'Transportation': totalExpenses * 0.25,
            'Shopping': totalExpenses * 0.15,
            'Entertainment': totalExpenses * 0.1,
            'Utilities': totalExpenses * 0.1
        };
        
        const categoryEntries = Object.entries(categories);

        mainContent.innerHTML = `
            <div class="categories-chart" style="grid-column: 1 / -1;">
                <div class="section-header">
                    <h2 class="section-title">Análisis por Categoría</h2>
                    <button class="btn btn-primary" onclick="loadDashboard()">
                        <i class="fas fa-arrow-left"></i> Volver al Dashboard
                    </button>
                </div>
                
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-top: 20px;">
                    <div>
                        <h3 style="margin-bottom: 20px; color: var(--dark-color);">Distribución de Gastos</h3>
                        <div style="height: 300px; display: flex; align-items: center; justify-content: center;">
                            <div style="width: 250px; height: 250px; border-radius: 50%; 
                                background: conic-gradient(${categoryEntries.map(([category, amount], index) => 
                                    `${getCategoryColor(category, index)} ${index === 0 ? '0%' : ''} ${(categoryEntries.slice(0, index + 1).reduce((sum, [_, amt]) => sum + (amt/totalExpenses * 100), 0))}%`
                                ).join(', ')}); 
                                display: flex; align-items: center; justify-content: center;">
                                <div style="background: white; width: 150px; height: 150px; border-radius: 50%; 
                                    display: flex; align-items: center; justify-content: center; flex-direction: column;">
                                    <div style="font-size: 24px; font-weight: bold; color: var(--dark-color);">
                                        $${totalExpenses.toFixed(2)}
                                    </div>
                                    <div style="font-size: 14px; color: var(--gray-color);">Total Gastos</div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div>
                        <h3 style="margin-bottom: 20px; color: var(--dark-color);">Detalles por Categoría</h3>
                        <div style="display: flex; flex-direction: column; gap: 15px;">
                            ${categoryEntries.length > 0 ? 
                                categoryEntries.map(([category, amount], index) => {
                                    const percentage = ((amount / totalExpenses) * 100).toFixed(1);
                                    return `
                                        <div class="category-item">
                                            <div class="category-info">
                                                <div class="category-color" style="background-color: ${getCategoryColor(category, index)}"></div>
                                                <div>
                                                    <div class="category-name">${category}</div>
                                                    <div style="font-size: 12px; color: var(--gray-color);">${percentage}% del total</div>
                                                </div>
                                            </div>
                                            <div>
                                                <div class="category-percentage">$${amount.toFixed(2)}</div>
                                                <div style="width: 150px; height: 8px; background: #eee; border-radius: 4px; margin-top: 5px;">
                                                    <div style="width: ${percentage}%; height: 100%; background-color: ${getCategoryColor(category, index)}; border-radius: 4px;"></div>
                                                </div>
                                            </div>
                                        </div>
                                    `;
                                }).join('') :
                                '<p style="text-align: center; color: var(--gray-color); padding: 40px;">No hay datos de categorías</p>'
                            }
                        </div>
                    </div>
                </div>
                
                <div style="margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 8px;">
                    <h3 style="margin-bottom: 15px; color: var(--dark-color);">Recomendaciones</h3>
                    ${categoryEntries.length > 0 ? `
                        <p>Basado en tus gastos, aquí tienes algunas recomendaciones:</p>
                        <ul style="margin-top: 10px; padding-left: 20px;">
                            <li>Tu categoría más alta es <strong>${Object.entries(categories).reduce((a, b) => a[1] > b[1] ? a : b)[0]}</strong></li>
                            <li>Considera establecer un presupuesto para las categorías con mayor gasto</li>
                            <li>Revisa tus gastos recurrentes en "Servicios" y "Transporte"</li>
                        </ul>
                    ` : '<p>Agrega más transacciones para obtener recomendaciones personalizadas.</p>'}
                </div>
            </div>
        `;

    } catch (error) {
        console.error('Error loading categories:', error);
        mainContent.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 20px; border-radius: 8px; text-align: center;">
                <h3><i class="fas fa-exclamation-circle"></i> Error al cargar categorías</h3>
                <p>No se pudo conectar con el servidor.</p>
                <button class="btn btn-primary" onclick="loadCategories()" style="margin-top: 15px;">
                    <i class="fas fa-redo"></i> Reintentar
                </button>
            </div>
        `;
    }
}

function loadUpload() {
    mainContent.innerHTML = `
        <div class="upload-section" style="grid-column: 1 / -1;">
            <div class="section-header">
                <h2 class="section-title">Subir Resumen de Tarjeta</h2>
                <button class="btn btn-primary" onclick="loadDashboard()">
                    <i class="fas fa-arrow-left"></i> Volver al Dashboard
                </button>
            </div>
            
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-top: 20px;">
                <div>
                    <h3 style="margin-bottom: 15px; color: var(--dark-color);">Cómo funciona</h3>
                    <div style="background: white; padding: 25px; border-radius: 12px; box-shadow: var(--box-shadow);">
                        <ol style="padding-left: 20px; line-height: 2;">
                            <li>Descarga el resumen de tu tarjeta de crédito en formato PDF</li>
                            <li>Sube el archivo usando el botón de abajo</li>
                            <li>Nuestro sistema extraerá automáticamente las transacciones</li>
                            <li>Revisa y confirma las transacciones detectadas</li>
                            <li>¡Listo! Los gastos se agregarán a tu dashboard</li>
                        </ol>
                    </div>
                    
                    <div style="background: #e8f4fd; padding: 20px; border-radius: 12px; margin-top: 20px; border-left: 4px solid var(--primary-color);">
                        <h4 style="color: var(--primary-color); margin-bottom: 10px;">
                            <i class="fas fa-lightbulb"></i> Consejo
                        </h4>
                        <p style="color: var(--dark-color);">
                            Asegúrate de que el PDF sea legible y contenga información de transacciones. 
                            El sistema funciona mejor con resúmenes de bancos principales.
                        </p>
                    </div>
                </div>
                
                <div>
                    <div class="upload-area" style="height: 300px; display: flex; flex-direction: column; justify-content: center; align-items: center; cursor: pointer;" 
                         onclick="uploadModal.classList.add('active')">
                        <div class="upload-icon">
                            <i class="fas fa-file-pdf"></i>
                        </div>
                        <h3>Haz clic para subir PDF</h3>
                        <p class="upload-text">o arrastra y suelta el archivo aquí</p>
                        <button class="upload-button">
                            Seleccionar Archivo
                        </button>
                        <p style="margin-top: 15px; font-size: 14px; color: var(--gray-color);">
                            Formatos soportados: PDF (máx. 10MB)
                        </p>
                    </div>
                    
                    <div style="background: white; padding: 20px; border-radius: 12px; margin-top: 20px; box-shadow: var(--box-shadow);">
                        <h4 style="margin-bottom: 15px; color: var(--dark-color);">
                            <i class="fas fa-history"></i> Historial de Subidas
                        </h4>
                        <p style="color: var(--gray-color); text-align: center; padding: 20px;">
                            No hay archivos subidos recientemente
                        </p>
                    </div>
                </div>
            </div>
        </div>
    `;
}

function loadAlerts() {
    mainContent.innerHTML = `
        <div class="alerts-section" style="grid-column: 1 / -1;">
            <div class="section-header">
                <h2 class="section-title">Alertas y Notificaciones</h2>
                <button class="btn btn-primary" onclick="loadDashboard()">
                    <i class="fas fa-arrow-left"></i> Volver al Dashboard
                </button>
            </div>
            
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 25px; margin-top: 20px;">
                <div>
                    <h3 style="margin-bottom: 20px; color: var(--dark-color);">Alertas Activas</h3>
                    
                    <div class="alert-item">
                        <div class="alert-icon" style="color: #e84118;">
                            <i class="fas fa-exclamation-triangle"></i>
                        </div>
                        <div class="alert-content">
                            <h4>Gasto alto en Comida</h4>
                            <p>Has gastado $150 en comida este mes, superando tu presupuesto de $100</p>
                            <small style="color: var(--gray-color);">Hace 2 días</small>
                        </div>
                    </div>
                    
                    <div class="alert-item">
                        <div class="alert-icon" style="color: #fbc531;">
                            <i class="fas fa-info-circle"></i>
                        </div>
                        <div class="alert-content">
                            <h4>Resumen pendiente</h4>
                            <p>Tu resumen de tarjeta de crédito de Mayo está listo para revisión</p>
                            <small style="color: var(--gray-color);">Hace 5 días</small>
                        </div>
                    </div>
                    
                    <div class="alert-item">
                        <div class="alert-icon" style="color: #00a8ff;">
                            <i class="fas fa-bell"></i>
                        </div>
                        <div class="alert-content">
                            <h4>Pago de tarjeta próximo</h4>
                            <p>Tu pago de tarjeta de crédito vence en 3 días</p>
                            <small style="color: var(--gray-color);">Hace 1 semana</small>
                        </div>
                    </div>
                </div>
                
                <div>
                    <h3 style="margin-bottom: 20px; color: var(--dark-color);">Configurar Alertas</h3>
                    
                    <div style="background: white; padding: 25px; border-radius: 12px; box-shadow: var(--box-shadow);">
                        <div class="form-group">
                            <label class="form-label">Presupuesto mensual para Comida</label>
                            <input type="number" class="form-input" placeholder="Ej: 300" value="100">
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Presupuesto mensual para Transporte</label>
                            <input type="number" class="form-input" placeholder="Ej: 150" value="80">
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Presupuesto mensual para Compras</label>
                            <input type="number" class="form-input" placeholder="Ej: 200" value="150">
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Notificar cuando el gasto supere</label>
                            <select class="form-select">
                                <option>80% del presupuesto</option>
                                <option selected>90% del presupuesto</option>
                                <option>100% del presupuesto</option>
                                <option>110% del presupuesto</option>
                            </select>
                        </div>
                        
                        <button class="btn btn-primary" style="width: 100%; padding: 15px; margin-top: 20px;">
                            <i class="fas fa-save"></i> Guardar Configuración
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
}

function loadSettings() {
    mainContent.innerHTML = `
        <div style="background: white; border-radius: var(--border-radius); padding: 30px; box-shadow: var(--box-shadow); grid-column: 1 / -1;">
            <div class="section-header">
                <h2 class="section-title">Configuración</h2>
                <button class="btn btn-primary" onclick="loadDashboard()">
                    <i class="fas fa-arrow-left"></i> Volver al Dashboard
                </button>
            </div>
            
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-top: 20px;">
                <div>
                    <h3 style="margin-bottom: 20px; color: var(--dark-color);">Preferencias</h3>
                    
                    <div style="background: #f8f9fa; padding: 25px; border-radius: 12px;">
                        <div class="form-group">
                            <label class="form-label">Moneda predeterminada</label>
                            <select class="form-select">
                                <option selected>USD - Dólar estadounidense</option>
                                <option>EUR - Euro</option>
                                <option>GBP - Libra esterlina</option>
                                <option>JPY - Yen japonés</option>
                                <option>CNY - Yuan chino</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Formato de fecha</label>
                            <select class="form-select">
                                <option selected>DD/MM/YYYY</option>
                                <option>MM/DD/YYYY</option>
                                <option>YYYY-MM-DD</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Notificaciones por email</label>
                            <div style="display: flex; align-items: center; gap: 10px; margin-top: 10px;">
                                <input type="checkbox" id="emailNotifications" checked>
                                <label for="emailNotifications">Recibir resumen semanal por email</label>
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label class="form-label">Notificaciones push</label>
                            <div style="display: flex; align-items: center; gap: 10px; margin-top: 10px;">
                                <input type="checkbox" id="pushNotifications" checked>
                                <label for="pushNotifications">Alertas de gastos altos</label>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div>
                    <h3 style="margin-bottom: 20px; color: var(--dark-color);">Categorías Personalizadas</h3>
                    
                    <div style="background: #f8f9fa; padding: 25px; border-radius: 12px;">
                        <div id="customCategories">
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                                <span>Alimentos</span>
                                <button class="btn" style="padding: 5px 10px; font-size: 12px; background: var(--danger-color); color: white;">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                                <span>Transporte</span>
                                <button class="btn" style="padding: 5px 10px; font-size: 12px; background: var(--danger-color); color: white;">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                                <span>Compras</span>
                                <button class="btn" style="padding: 5px 10px; font-size: 12px; background: var(--danger-color); color: white;">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                                <span>Entretenimiento</span>
                                <button class="btn" style="padding: 5px 10px; font-size: 12px; background: var(--danger-color); color: white;">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                        
                        <div style="display: flex; gap: 10px; margin-top: 20px;">
                            <input type="text" class="form-input" id="newCategory" placeholder="Nueva categoría" style="flex: 1;">
                            <button class="btn btn-success">
                                <i class="fas fa-plus"></i> Agregar
                            </button>
                        </div>
                    </div>
                    
                    <div style="margin-top: 30px;">
                        <h3 style="margin-bottom: 15px; color: var(--dark-color);">Exportar Datos</h3>
                        <div style="display: flex; gap: 10px;">
                            <button class="btn" style="flex: 1; background: var(--primary-color); color: white;">
                                <i class="fas fa-file-csv"></i> CSV
                            </button>
                            <button class="btn" style="flex: 1; background: var(--secondary-color); color: white;">
                                <i class="fas fa-file-excel"></i> Excel
                            </button>
                            <button class="btn" style="flex: 1; background: var(--success-color); color: white;">
                                <i class="fas fa-file-pdf"></i> PDF
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            
            <div style="margin-top: 30px; display: flex; justify-content: flex-end; gap: 15px;">
                <button class="btn" style="background: var(--gray-color); color: white;">
                    Cancelar
                </button>
                <button class="btn btn-primary">
                    <i class="fas fa-save"></i> Guardar Cambios
                </button>
            </div>
        </div>
    `;
}

async function loadCirclesPage() {
    try {
        const result = await auth.getCircles();
        if (!result.success) throw new Error(result.error || 'Error al cargar círculos');
        
        const circles = result.data?.circles || [];
        userCircles = circles;

        mainContent.innerHTML = `
            <div style="background: white; border-radius: var(--border-radius); padding: 30px; box-shadow: var(--box-shadow); grid-column: 1 / -1;">
                <div class="section-header">
                    <h2 class="section-title">Círculos Familiares</h2>
                    <div>
                        <button class="btn btn-primary" onclick="createCircle()">
                            <i class="fas fa-plus"></i> Crear Círculo
                        </button>
                        <button class="btn btn-success" onclick="joinCircle()" style="margin-left: 10px;">
                            <i class="fas fa-user-plus"></i> Unirse a Círculo
                        </button>
                    </div>
                </div>
                
                <div style="margin-top: 20px;">
                    ${circles.length > 0 ? circles.map(circle => `
                        <div class="circle-card" style="background: #f8f9fa; border-radius: 8px; padding: 20px; margin-bottom: 15px; border-left: 4px solid var(--primary-color);">
                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                <div>
                                    <h3 style="margin: 0 0 5px 0; color: var(--dark-color);">${circle.name}</h3>
                                    <p style="margin: 0; color: var(--gray-color); font-size: 14px;">${circle.description || 'Sin descripción'}</p>
                                    <div style="display: flex; gap: 10px; margin-top: 10px;">
                                        <span style="background: var(--primary-color); color: white; padding: 3px 8px; border-radius: 4px; font-size: 12px;">
                                            ${circle.currency || 'USD'}
                                        </span>
                                        <span style="background: ${circle.isPublic ? '#28a745' : '#6c757d'}; color: white; padding: 3px 8px; border-radius: 4px; font-size: 12px;">
                                            ${circle.isPublic ? 'Público' : 'Privado'}
                                        </span>
                                    </div>
                                </div>
                                <div>
                                    <button class="btn" onclick="switchCircle('${circle.id}')" style="margin-right: 10px;">
                                        <i class="fas fa-eye"></i> Ver
                                    </button>
                                    <button class="btn" style="background: var(--warning-color); color: white;">
                                        <i class="fas fa-cog"></i> Configurar
                                    </button>
                                </div>
                            </div>
                        </div>
                    `).join('') : `
                        <div style="text-align: center; padding: 40px; color: var(--gray-color);">
                            <i class="fas fa-users" style="font-size: 48px; margin-bottom: 20px; color: #ddd;"></i>
                            <h3>No tienes círculos familiares</h3>
                            <p>Crea tu primer círculo para compartir gastos con familiares o amigos.</p>
                            <button class="btn btn-primary" onclick="createCircle()" style="margin-top: 20px;">
                                <i class="fas fa-plus"></i> Crear Primer Círculo
                            </button>
                        </div>
                    `}
                </div>
                
                <div style="margin-top: 30px; padding: 20px; background: #e8f4fd; border-radius: 8px; border-left: 4px solid var(--info-color);">
                    <h4 style="color: var(--info-color); margin-bottom: 10px;">
                        <i class="fas fa-lightbulb"></i> ¿Qué son los círculos familiares?
                    </h4>
                    <p style="color: var(--dark-color); margin-bottom: 10px;">
                        Los círculos familiares te permiten compartir gastos e ingresos con tu familia o grupo de amigos.
                        Cada círculo tiene su propio balance y transacciones compartidas.
                    </p>
                    <ul style="color: var(--dark-color); padding-left: 20px;">
                        <li><strong>Creador/Owner:</strong> Puede administrar miembros y eliminar el círculo</li>
                        <li><strong>Admin:</strong> Puede agregar/eliminar miembros y aprobar transacciones</li>
                        <li><strong>Member:</strong> Puede agregar transacciones y ver el balance</li>
                        <li><strong>Viewer:</strong> Solo puede ver transacciones (no puede agregar)</li>
                    </ul>
                </div>
            </div>
        `;
    } catch (error) {
        console.error('Error loading circles:', error);
        mainContent.innerHTML = `
            <div style="background: #f8d7da; color: #721c24; padding: 20px; border-radius: 8px; text-align: center;">
                <h3><i class="fas fa-exclamation-circle"></i> Error al cargar círculos</h3>
                <p>${error.message}</p>
                <button class="btn btn-primary" onclick="loadCirclesPage()" style="margin-top: 15px;">
                    <i class="fas fa-redo"></i> Reintentar
                </button>
            </div>
        `;
    }
}

// Helper functions
function getCategoryClass(category) {
    const classes = {
        'Food': 'food',
        'Transportation': 'transport',
        'Shopping': 'shopping',
        'Entertainment': 'entertainment',
        'Coffee': 'coffee',
        'Utilities': 'transport',
        'Healthcare': 'food',
        'Education': 'entertainment',
        'Other': 'shopping'
    };
    return classes[category] || 'shopping';
}

function getCategoryIcon(category) {
    const icons = {
        'Food': 'fas fa-utensils',
        'Transportation': 'fas fa-car',
        'Shopping': 'fas fa-shopping-bag',
        'Entertainment': 'fas fa-film',
        'Coffee': 'fas fa-coffee',
        'Utilities': 'fas fa-bolt',
        'Healthcare': 'fas fa-heartbeat',
        'Education': 'fas fa-graduation-cap',
        'Other': 'fas fa-question-circle'
    };
    return icons[category] || 'fas fa-question-circle';
}

function getCategoryColor(category, index) {
    const colors = [
        '#FF6B6B', '#4ECDC4', '#FFD166', '#06D6A0', 
        '#118AB2', '#EF476F', '#073B4C', '#7209B7',
        '#F72585', '#3A0CA3', '#4361EE', '#4CC9F0'
    ];
    return colors[index % colors.length];
}

function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString('es-ES', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric'
        });
    } catch (e) {
        return 'Fecha inválida';
    }
}

// Load user circles
async function loadUserCircles() {
    try {
        const result = await auth.getCircles();
        if (result.success) {
            userCircles = result.data?.circles || [];
            updateCircleSelector();
        }
    } catch (error) {
        console.error('Error loading circles:', error);
    }
}

// Update circle selector in forms
function updateCircleSelector() {
    const circleSelect = document.getElementById('circle');
    if (!circleSelect) return;
    
    // Clear existing options
    circleSelect.innerHTML = '<option value="">Personal (solo yo)</option>';
    
    // Add circle options
    userCircles.forEach(circle => {
        const option = document.createElement('option');
        option.value = circle.id;
        option.textContent = circle.name;
        circleSelect.appendChild(option);
    });
}

// Create new circle
async function createCircle() {
    const name = prompt('Nombre del círculo familiar:');
    if (!name) return;
    
    const description = prompt('Descripción (opcional):');
    const currency = prompt('Moneda (USD, EUR, etc.):', 'USD');
    
    try {
        const result = await auth.createCircle(name, description, currency);
        if (result.success) {
            alert('Círculo creado exitosamente!');
            await loadUserCircles();
            loadCirclesPage();
        } else {
            alert('Error: ' + result.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

// Join circle with code
async function joinCircle() {
    const code = prompt('Código de invitación:');
    if (!code) return;
    
    try {
        const result = await auth.joinCircle(code);
        if (result.success) {
            alert('Te has unido al círculo exitosamente!');
            await loadUserCircles();
            loadCirclesPage();
        } else {
            alert('Error: ' + result.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

// Switch current circle
function switchCircle(circleId) {
    currentCircle = circleId || null;
    loadDashboard();
}

// Update auth UI
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

async function checkAPIConnectivity() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        if (!response.ok) {
            console.warn('MCP Server API not responding');
        }
    } catch (error) {
        console.warn('Cannot connect to MCP Server API:', error);
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', initApp);

// Make functions available globally for onclick handlers
window.loadPage = loadPage;
window.loadDashboard = loadDashboard;
window.loadTransactions = loadTransactions;
window.loadCategories = loadCategories;
window.loadUpload = loadUpload;
window.loadAlerts = loadAlerts;
window.loadSettings = loadSettings;
window.loadCirclesPage = loadCirclesPage;
window.createCircle = createCircle;
window.joinCircle = joinCircle;
window.switchCircle = switchCircle;
window.handleLogout = handleLogout;
window.auth = auth;