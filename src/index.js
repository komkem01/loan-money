require('dotenv').config();
const express = require('express');
const cors = require('cors');
const db = require('./database/db');
const authHandler = require('./handlers/auth');
const profileHandler = require('./handlers/profile');
const dashboardHandler = require('./handlers/dashboard');
const loanHandler = require('./handlers/loan');
const transactionHandler = require('./handlers/transaction');
const { authMiddleware } = require('./middleware/auth');
const { respondWithJSON, logAPICall } = require('./utils/response');

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors({
  origin: process.env.FRONTEND_URL || '*',
  credentials: true
}));
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Logging middleware
app.use((req, res, next) => {
  res.on('finish', () => {
    const userId = req.user ? req.user.id : 'anonymous';
    logAPICall(req.method, req.path, userId, res.statusCode);
  });
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  respondWithJSON(res, 200, { 
    status: 'healthy',
    timestamp: new Date().toISOString(),
    service: 'loan-money-api'
  });
});

// Auth routes (public)
app.post('/api/v1/register', authHandler.register.bind(authHandler));
app.post('/api/v1/login', authHandler.login.bind(authHandler));

// Protected routes (authentication required)
app.use('/api/v1', authMiddleware);

// Profile management endpoints
app.get('/api/v1/profile', profileHandler.getProfile.bind(profileHandler));
app.patch('/api/v1/profile', profileHandler.updateProfile.bind(profileHandler));
app.patch('/api/v1/change-password', profileHandler.changePassword.bind(profileHandler));

// Dashboard endpoints
app.get('/api/v1/dashboard/stats', dashboardHandler.getDashboardStats.bind(dashboardHandler));
app.get('/api/v1/dashboard/recent-transactions', dashboardHandler.getRecentTransactions.bind(dashboardHandler));
app.get('/api/v1/dashboard/loan-summary', dashboardHandler.getLoanSummary.bind(dashboardHandler));
app.get('/api/v1/dashboard/monthly-stats', dashboardHandler.getMonthlyStats.bind(dashboardHandler));
app.get('/api/v1/dashboard/overdue-loans', dashboardHandler.getOverdueLoans.bind(dashboardHandler));

// Loan management endpoints
app.get('/api/v1/loans', loanHandler.getLoans.bind(loanHandler));
app.post('/api/v1/loans', loanHandler.createLoan.bind(loanHandler));
app.get('/api/v1/loans/:id', loanHandler.getLoan.bind(loanHandler));
app.patch('/api/v1/loans/:id', loanHandler.updateLoan.bind(loanHandler));
app.delete('/api/v1/loans/:id', loanHandler.deleteLoan.bind(loanHandler));
app.patch('/api/v1/loans/:id/status', loanHandler.updateLoanStatus.bind(loanHandler));

// Transaction management endpoints
app.get('/api/v1/transactions', transactionHandler.getTransactions.bind(transactionHandler));
app.post('/api/v1/transactions', transactionHandler.createTransaction.bind(transactionHandler));
app.get('/api/v1/transactions/:id', transactionHandler.getTransaction.bind(transactionHandler));
app.patch('/api/v1/transactions/:id', transactionHandler.updateTransaction.bind(transactionHandler));
app.delete('/api/v1/transactions/:id', transactionHandler.deleteTransaction.bind(transactionHandler));
app.get('/api/v1/loans/:loanId/transactions', transactionHandler.getTransactionsByLoan.bind(transactionHandler));

// Error handling middleware
app.use((error, req, res, next) => {
  console.error('Unhandled error:', error);
  res.status(500).json({
    error: {
      message: 'Internal server error',
      status: 500
    }
  });
});

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: {
      message: 'Route not found',
      status: 404
    }
  });
});

// Initialize database and start server
async function startServer() {
  try {
    // Create database tables if they don't exist
    await db.createTables();
    console.log('Database initialized successfully');

    app.listen(PORT, () => {
      console.log(`Server running on port ${PORT}`);
      console.log(`Health check: http://localhost:${PORT}/health`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

// For Vercel serverless deployment
if (process.env.VERCEL) {
  module.exports = app;
} else {
  startServer();
}