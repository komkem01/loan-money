const db = require('../database/db');
const { respondWithError, respondWithJSON, validateRequiredFields, parsePagination } = require('../utils/response');
const { getUserFromContext } = require('../middleware/auth');
const { Transaction } = require('../models');

class TransactionHandler {
  /**
   * Get all transactions for user
   */
  async getTransactions(req, res) {
    try {
      const user = getUserFromContext(req);
      const { page, limit, offset } = parsePagination(req.query);
      const { loanId, transactionType } = req.query;

      let query = `
        SELECT t.*, l.borrower_name, l.amount as loan_amount
        FROM transactions t
        JOIN loans l ON t.loan_id = l.id
        WHERE t.user_id = $1
      `;
      let params = [user.id];
      let paramCount = 1;

      if (loanId) {
        paramCount++;
        query += ` AND t.loan_id = $${paramCount}`;
        params.push(loanId);
      }

      if (transactionType) {
        paramCount++;
        query += ` AND t.transaction_type = $${paramCount}`;
        params.push(transactionType);
      }

      query += ' ORDER BY t.created_at DESC';

      if (limit) {
        paramCount++;
        query += ` LIMIT $${paramCount}`;
        params.push(limit);

        if (offset) {
          paramCount++;
          query += ` OFFSET $${paramCount}`;
          params.push(offset);
        }
      }

      const result = await db.query(query, params);

      return respondWithJSON(res, 200, {
        transactions: result.rows,
        pagination: { page, limit, total: result.rowCount }
      });

    } catch (error) {
      console.error('Get transactions error:', error);
      return respondWithError(res, 500, 'Failed to get transactions');
    }
  }

  /**
   * Create new transaction
   */
  async createTransaction(req, res) {
    try {
      const user = getUserFromContext(req);
      const { loanId, amount, transactionType, transactionDate, description } = req.body;

      validateRequiredFields(req.body, ['loanId', 'amount', 'transactionType', 'transactionDate']);

      if (amount <= 0) {
        return respondWithError(res, 400, 'Amount must be greater than 0');
      }

      const validTypes = ['payment', 'interest', 'fee', 'adjustment'];
      if (!validTypes.includes(transactionType)) {
        return respondWithError(res, 400, `Transaction type must be one of: ${validTypes.join(', ')}`);
      }

      // Verify loan belongs to user
      const loanCheck = await db.query(
        'SELECT id FROM loans WHERE id = $1 AND user_id = $2',
        [loanId, user.id]
      );

      if (loanCheck.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      const result = await db.query(
        `INSERT INTO transactions (loan_id, user_id, amount, transaction_type, transaction_date, description)
         VALUES ($1, $2, $3, $4, $5, $6)
         RETURNING *`,
        [loanId, user.id, amount, transactionType, transactionDate, description]
      );

      const transactionData = result.rows[0];
      const transaction = new Transaction({
        id: transactionData.id,
        loanId: transactionData.loan_id,
        userId: transactionData.user_id,
        amount: transactionData.amount,
        transactionType: transactionData.transaction_type,
        transactionDate: transactionData.transaction_date,
        description: transactionData.description,
        createdAt: transactionData.created_at,
        updatedAt: transactionData.updated_at
      });

      return respondWithJSON(res, 201, transaction);

    } catch (error) {
      console.error('Create transaction error:', error);
      return respondWithError(res, 500, 'Failed to create transaction');
    }
  }

  /**
   * Get specific transaction
   */
  async getTransaction(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;

      const result = await db.query(
        `SELECT t.*, l.borrower_name, l.amount as loan_amount
         FROM transactions t
         JOIN loans l ON t.loan_id = l.id
         WHERE t.id = $1 AND t.user_id = $2`,
        [id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Transaction not found');
      }

      return respondWithJSON(res, 200, result.rows[0]);

    } catch (error) {
      console.error('Get transaction error:', error);
      return respondWithError(res, 500, 'Failed to get transaction');
    }
  }

  /**
   * Update transaction
   */
  async updateTransaction(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;
      const { amount, transactionType, transactionDate, description } = req.body;

      // Check if transaction exists and belongs to user
      const existingTransaction = await db.query(
        'SELECT * FROM transactions WHERE id = $1 AND user_id = $2',
        [id, user.id]
      );

      if (existingTransaction.rows.length === 0) {
        return respondWithError(res, 404, 'Transaction not found');
      }

      const result = await db.query(
        `UPDATE transactions 
         SET amount = $1, transaction_type = $2, transaction_date = $3, 
             description = $4, updated_at = CURRENT_TIMESTAMP
         WHERE id = $5 AND user_id = $6
         RETURNING *`,
        [amount, transactionType, transactionDate, description, id, user.id]
      );

      return respondWithJSON(res, 200, result.rows[0]);

    } catch (error) {
      console.error('Update transaction error:', error);
      return respondWithError(res, 500, 'Failed to update transaction');
    }
  }

  /**
   * Delete transaction
   */
  async deleteTransaction(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;

      const result = await db.query(
        'DELETE FROM transactions WHERE id = $1 AND user_id = $2 RETURNING *',
        [id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Transaction not found');
      }

      return respondWithJSON(res, 200, { message: 'Transaction deleted successfully' });

    } catch (error) {
      console.error('Delete transaction error:', error);
      return respondWithError(res, 500, 'Failed to delete transaction');
    }
  }

  /**
   * Get transactions by loan ID
   */
  async getTransactionsByLoan(req, res) {
    try {
      const user = getUserFromContext(req);
      const { loanId } = req.params;
      const { page, limit, offset } = parsePagination(req.query);

      // Verify loan belongs to user
      const loanCheck = await db.query(
        'SELECT id FROM loans WHERE id = $1 AND user_id = $2',
        [loanId, user.id]
      );

      if (loanCheck.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      let query = 'SELECT * FROM transactions WHERE loan_id = $1 ORDER BY created_at DESC';
      let params = [loanId];

      if (limit) {
        query += ' LIMIT $2';
        params.push(limit);

        if (offset) {
          query += ' OFFSET $3';
          params.push(offset);
        }
      }

      const result = await db.query(query, params);

      return respondWithJSON(res, 200, {
        transactions: result.rows,
        pagination: { page, limit, total: result.rowCount }
      });

    } catch (error) {
      console.error('Get transactions by loan error:', error);
      return respondWithError(res, 500, 'Failed to get loan transactions');
    }
  }
}

module.exports = new TransactionHandler();