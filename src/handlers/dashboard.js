const db = require('../database/db');
const { respondWithError, respondWithJSON, parsePagination } = require('../utils/response');
const { getUserFromContext } = require('../middleware/auth');
const { DashboardStats } = require('../models');

class DashboardHandler {
  /**
   * Get dashboard statistics
   */
  async getDashboardStats(req, res) {
    try {
      const user = getUserFromContext(req);

      // Get total loans count
      const totalLoansResult = await db.query(
        'SELECT COUNT(*) as count FROM loans WHERE user_id = $1',
        [user.id]
      );

      // Get active loans count
      const activeLoansResult = await db.query(
        'SELECT COUNT(*) as count FROM loans WHERE user_id = $1 AND status = $2',
        [user.id, 'active']
      );

      // Get total amount
      const totalAmountResult = await db.query(
        'SELECT COALESCE(SUM(amount), 0) as total FROM loans WHERE user_id = $1',
        [user.id]
      );

      // Get overdue loans count
      const overdueLoansResult = await db.query(
        'SELECT COUNT(*) as count FROM loans WHERE user_id = $1 AND due_date < CURRENT_DATE AND status = $2',
        [user.id, 'active']
      );

      const stats = new DashboardStats({
        totalLoans: parseInt(totalLoansResult.rows[0].count),
        activeLoans: parseInt(activeLoansResult.rows[0].count),
        totalAmount: parseFloat(totalAmountResult.rows[0].total),
        totalInterest: 0, // Calculate based on business logic
        overdueLoans: parseInt(overdueLoansResult.rows[0].count)
      });

      return respondWithJSON(res, 200, stats);

    } catch (error) {
      console.error('Dashboard stats error:', error);
      return respondWithError(res, 500, 'Failed to get dashboard statistics');
    }
  }

  /**
   * Get recent transactions
   */
  async getRecentTransactions(req, res) {
    try {
      const user = getUserFromContext(req);
      const { limit } = parsePagination(req.query);

      const result = await db.query(
        `SELECT t.*, l.borrower_name, l.amount as loan_amount
         FROM transactions t
         JOIN loans l ON t.loan_id = l.id
         WHERE t.user_id = $1
         ORDER BY t.created_at DESC
         LIMIT $2`,
        [user.id, limit || 10]
      );

      return respondWithJSON(res, 200, result.rows);

    } catch (error) {
      console.error('Recent transactions error:', error);
      return respondWithError(res, 500, 'Failed to get recent transactions');
    }
  }

  /**
   * Get loan summary
   */
  async getLoanSummary(req, res) {
    try {
      const user = getUserFromContext(req);

      const result = await db.query(
        `SELECT 
           status,
           COUNT(*) as count,
           COALESCE(SUM(amount), 0) as total_amount
         FROM loans 
         WHERE user_id = $1 
         GROUP BY status`,
        [user.id]
      );

      return respondWithJSON(res, 200, result.rows);

    } catch (error) {
      console.error('Loan summary error:', error);
      return respondWithError(res, 500, 'Failed to get loan summary');
    }
  }

  /**
   * Get monthly statistics
   */
  async getMonthlyStats(req, res) {
    try {
      const user = getUserFromContext(req);

      const result = await db.query(
        `SELECT 
           DATE_TRUNC('month', loan_date) as month,
           COUNT(*) as loans_count,
           COALESCE(SUM(amount), 0) as total_amount
         FROM loans 
         WHERE user_id = $1 
         GROUP BY DATE_TRUNC('month', loan_date)
         ORDER BY month DESC
         LIMIT 12`,
        [user.id]
      );

      return respondWithJSON(res, 200, result.rows);

    } catch (error) {
      console.error('Monthly stats error:', error);
      return respondWithError(res, 500, 'Failed to get monthly statistics');
    }
  }

  /**
   * Get overdue loans
   */
  async getOverdueLoans(req, res) {
    try {
      const user = getUserFromContext(req);

      const result = await db.query(
        `SELECT * FROM loans 
         WHERE user_id = $1 
         AND due_date < CURRENT_DATE 
         AND status = 'active'
         ORDER BY due_date ASC`,
        [user.id]
      );

      return respondWithJSON(res, 200, result.rows);

    } catch (error) {
      console.error('Overdue loans error:', error);
      return respondWithError(res, 500, 'Failed to get overdue loans');
    }
  }
}

module.exports = new DashboardHandler();