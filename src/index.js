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
  origin: true, // Allow all origins including file://
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With']
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

// Auth routes (public) - NO AUTH REQUIRED
app.post('/api/v1/register', authHandler.register.bind(authHandler));
app.post('/api/v1/login', authHandler.login.bind(authHandler));

// Apply auth middleware only to protected routes
// Don't use app.use('/api/v1', authMiddleware) as it affects register/login too

// Profile management endpoints (protected)
app.get('/api/v1/profile', authMiddleware, profileHandler.getProfile.bind(profileHandler));
app.patch('/api/v1/profile', authMiddleware, profileHandler.updateProfile.bind(profileHandler));
app.patch('/api/v1/change-password', authMiddleware, profileHandler.changePassword.bind(profileHandler));

// Dashboard endpoints (protected)
app.get('/api/v1/dashboard/stats', authMiddleware, dashboardHandler.getDashboardStats.bind(dashboardHandler));
app.get('/api/v1/dashboard/recent-transactions', authMiddleware, dashboardHandler.getRecentTransactions.bind(dashboardHandler));
app.get('/api/v1/dashboard/loan-summary', authMiddleware, dashboardHandler.getLoanSummary.bind(dashboardHandler));
app.get('/api/v1/dashboard/monthly-stats', authMiddleware, dashboardHandler.getMonthlyStats.bind(dashboardHandler));
app.get('/api/v1/dashboard/overdue-loans', authMiddleware, dashboardHandler.getOverdueLoans.bind(dashboardHandler));

// Loan management endpoints (protected)
app.get('/api/v1/loans', authMiddleware, loanHandler.getLoans.bind(loanHandler));
app.post('/api/v1/loans', authMiddleware, loanHandler.createLoan.bind(loanHandler));
app.get('/api/v1/loans/:id', authMiddleware, loanHandler.getLoan.bind(loanHandler));
app.patch('/api/v1/loans/:id', authMiddleware, loanHandler.updateLoan.bind(loanHandler));
app.delete('/api/v1/loans/:id', authMiddleware, loanHandler.deleteLoan.bind(loanHandler));
app.patch('/api/v1/loans/:id/status', authMiddleware, loanHandler.updateLoanStatus.bind(loanHandler));

// Transaction management endpoints (protected)
app.get('/api/v1/transactions', authMiddleware, transactionHandler.getTransactions.bind(transactionHandler));
app.post('/api/v1/transactions', authMiddleware, transactionHandler.createTransaction.bind(transactionHandler));
app.get('/api/v1/transactions/:id', authMiddleware, transactionHandler.getTransaction.bind(transactionHandler));
app.patch('/api/v1/transactions/:id', authMiddleware, transactionHandler.updateTransaction.bind(transactionHandler));
app.delete('/api/v1/transactions/:id', authMiddleware, transactionHandler.deleteTransaction.bind(transactionHandler));
app.get('/api/v1/loans/:loanId/transactions', authMiddleware, transactionHandler.getTransactionsByLoan.bind(transactionHandler));

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