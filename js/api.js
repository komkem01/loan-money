// API Configuration
const API_BASE_URL = window.location.origin;

// Utility functions for API calls
class ApiClient {
  constructor() {
    this.token = localStorage.getItem('token');
    this.baseURL = `${API_BASE_URL}/api/v1`;
  }

  // Set authentication token
  setToken(token) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  // Get authentication headers
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json'
    };
    
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    
    return headers;
  }

  // Generic API call method
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      headers: this.getHeaders(),
      ...options
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error?.message || 'API request failed');
      }

      return data;
    } catch (error) {
      console.error('API Error:', error);
      
      // Handle authentication errors
      if (error.message.includes('token') || error.message.includes('unauthorized')) {
        this.setToken(null);
        window.location.href = '/index.html';
      }
      
      throw error;
    }
  }

  // Authentication methods
  async register(userData) {
    return this.request('/register', {
      method: 'POST',
      body: JSON.stringify(userData)
    });
  }

  async login(credentials) {
    return this.request('/login', {
      method: 'POST',
      body: JSON.stringify(credentials)
    });
  }

  // Profile methods
  async getProfile() {
    return this.request('/profile');
  }

  async updateProfile(profileData) {
    return this.request('/profile', {
      method: 'PATCH',
      body: JSON.stringify(profileData)
    });
  }

  async changePassword(passwordData) {
    return this.request('/change-password', {
      method: 'PATCH',
      body: JSON.stringify(passwordData)
    });
  }

  // Dashboard methods
  async getDashboardStats() {
    return this.request('/dashboard/stats');
  }

  async getRecentTransactions() {
    return this.request('/dashboard/recent-transactions');
  }

  async getLoanSummary() {
    return this.request('/dashboard/loan-summary');
  }

  async getMonthlyStats() {
    return this.request('/dashboard/monthly-stats');
  }

  async getOverdueLoans() {
    return this.request('/dashboard/overdue-loans');
  }

  // Loan methods
  async getLoans(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return this.request(`/loans${queryString ? '?' + queryString : ''}`);
  }

  async createLoan(loanData) {
    return this.request('/loans', {
      method: 'POST',
      body: JSON.stringify(loanData)
    });
  }

  async getLoan(id) {
    return this.request(`/loans/${id}`);
  }

  async updateLoan(id, loanData) {
    return this.request(`/loans/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(loanData)
    });
  }

  async deleteLoan(id) {
    return this.request(`/loans/${id}`, {
      method: 'DELETE'
    });
  }

  async updateLoanStatus(id, status) {
    return this.request(`/loans/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    });
  }

  // Transaction methods
  async getTransactions(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return this.request(`/transactions${queryString ? '?' + queryString : ''}`);
  }

  async createTransaction(transactionData) {
    return this.request('/transactions', {
      method: 'POST',
      body: JSON.stringify(transactionData)
    });
  }

  async getTransaction(id) {
    return this.request(`/transactions/${id}`);
  }

  async updateTransaction(id, transactionData) {
    return this.request(`/transactions/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(transactionData)
    });
  }

  async deleteTransaction(id) {
    return this.request(`/transactions/${id}`, {
      method: 'DELETE'
    });
  }

  async getTransactionsByLoan(loanId) {
    return this.request(`/loans/${loanId}/transactions`);
  }
}

// Create global API client instance
window.api = new ApiClient();

// Utility functions
function formatCurrency(amount) {
  return new Intl.NumberFormat('th-TH', {
    style: 'currency',
    currency: 'THB'
  }).format(amount);
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('th-TH');
}

function showNotification(message, type = 'info') {
  // Simple notification system
  const notification = document.createElement('div');
  notification.className = `notification ${type}`;
  notification.textContent = message;
  
  // Add some basic styling
  notification.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    padding: 10px 20px;
    border-radius: 5px;
    color: white;
    z-index: 1000;
    background-color: ${type === 'error' ? '#e74c3c' : type === 'success' ? '#2ecc71' : '#3498db'};
  `;
  
  document.body.appendChild(notification);
  
  setTimeout(() => {
    document.body.removeChild(notification);
  }, 3000);
}

// Authentication check
function checkAuth() {
  const token = localStorage.getItem('token');
  const currentPage = window.location.pathname;
  
  // Public pages that don't require authentication
  const publicPages = ['/index.html', '/register.html', '/'];
  
  if (!token && !publicPages.includes(currentPage)) {
    window.location.href = '/index.html';
    return false;
  }
  
  if (token && publicPages.includes(currentPage)) {
    window.location.href = '/dashboard.html';
    return false;
  }
  
  return true;
}

// Logout function
function logout() {
  window.api.setToken(null);
  window.location.href = '/index.html';
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', checkAuth);