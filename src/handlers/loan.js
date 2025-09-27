const db = require('../database/db');
const { respondWithError, respondWithJSON, validateRequiredFields, parsePagination } = require('../utils/response');
const { getUserFromContext } = require('../middleware/auth');
const { Loan } = require('../models');

class LoanHandler {
  /**
   * Get all loans for user
   */
  async getLoans(req, res) {
    try {
      const user = getUserFromContext(req);
      const { page, limit, offset } = parsePagination(req.query);
      const { status, search } = req.query;

      let query = 'SELECT * FROM loans WHERE user_id = $1';
      let params = [user.id];
      let paramCount = 1;

      if (status) {
        paramCount++;
        query += ` AND status = $${paramCount}`;
        params.push(status);
      }

      if (search) {
        paramCount++;
        query += ` AND borrower_name ILIKE $${paramCount}`;
        params.push(`%${search}%`);
      }

      query += ' ORDER BY created_at DESC';

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
        loans: result.rows,
        pagination: { page, limit, total: result.rowCount }
      });

    } catch (error) {
      console.error('Get loans error:', error);
      return respondWithError(res, 500, 'Failed to get loans');
    }
  }

  /**
   * Create new loan
   */
  async createLoan(req, res) {
    try {
      const user = getUserFromContext(req);
      const { borrowerName, borrowerPhone, borrowerAddress, amount, interestRate, loanDate, dueDate, notes } = req.body;

      validateRequiredFields(req.body, ['borrowerName', 'amount', 'interestRate', 'loanDate']);

      if (amount <= 0) {
        return respondWithError(res, 400, 'Amount must be greater than 0');
      }

      if (interestRate < 0) {
        return respondWithError(res, 400, 'Interest rate cannot be negative');
      }

      const result = await db.query(
        `INSERT INTO loans (user_id, borrower_name, borrower_phone, borrower_address, amount, interest_rate, loan_date, due_date, notes)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING *`,
        [user.id, borrowerName, borrowerPhone, borrowerAddress, amount, interestRate, loanDate, dueDate, notes]
      );

      const loanData = result.rows[0];
      const loan = new Loan({
        id: loanData.id,
        userId: loanData.user_id,
        borrowerName: loanData.borrower_name,
        borrowerPhone: loanData.borrower_phone,
        borrowerAddress: loanData.borrower_address,
        amount: loanData.amount,
        interestRate: loanData.interest_rate,
        loanDate: loanData.loan_date,
        dueDate: loanData.due_date,
        status: loanData.status,
        notes: loanData.notes,
        createdAt: loanData.created_at,
        updatedAt: loanData.updated_at
      });

      return respondWithJSON(res, 201, loan);

    } catch (error) {
      console.error('Create loan error:', error);
      return respondWithError(res, 500, 'Failed to create loan');
    }
  }

  /**
   * Get specific loan
   */
  async getLoan(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;

      const result = await db.query(
        'SELECT * FROM loans WHERE id = $1 AND user_id = $2',
        [id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      return respondWithJSON(res, 200, result.rows[0]);

    } catch (error) {
      console.error('Get loan error:', error);
      return respondWithError(res, 500, 'Failed to get loan');
    }
  }

  /**
   * Update loan
   */
  async updateLoan(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;
      const { borrowerName, borrowerPhone, borrowerAddress, amount, interestRate, loanDate, dueDate, notes } = req.body;

      const result = await db.query(
        `UPDATE loans 
         SET borrower_name = $1, borrower_phone = $2, borrower_address = $3, 
             amount = $4, interest_rate = $5, loan_date = $6, due_date = $7, 
             notes = $8, updated_at = CURRENT_TIMESTAMP
         WHERE id = $9 AND user_id = $10
         RETURNING *`,
        [borrowerName, borrowerPhone, borrowerAddress, amount, interestRate, loanDate, dueDate, notes, id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      return respondWithJSON(res, 200, result.rows[0]);

    } catch (error) {
      console.error('Update loan error:', error);
      return respondWithError(res, 500, 'Failed to update loan');
    }
  }

  /**
   * Delete loan
   */
  async deleteLoan(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;

      const result = await db.query(
        'DELETE FROM loans WHERE id = $1 AND user_id = $2 RETURNING *',
        [id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      return respondWithJSON(res, 200, { message: 'Loan deleted successfully' });

    } catch (error) {
      console.error('Delete loan error:', error);
      return respondWithError(res, 500, 'Failed to delete loan');
    }
  }

  /**
   * Update loan status
   */
  async updateLoanStatus(req, res) {
    try {
      const user = getUserFromContext(req);
      const { id } = req.params;
      const { status } = req.body;

      validateRequiredFields(req.body, ['status']);

      const validStatuses = ['active', 'paid', 'overdue', 'defaulted'];
      if (!validStatuses.includes(status)) {
        return respondWithError(res, 400, `Status must be one of: ${validStatuses.join(', ')}`);
      }

      const result = await db.query(
        'UPDATE loans SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3 RETURNING *',
        [status, id, user.id]
      );

      if (result.rows.length === 0) {
        return respondWithError(res, 404, 'Loan not found');
      }

      return respondWithJSON(res, 200, result.rows[0]);

    } catch (error) {
      console.error('Update loan status error:', error);
      return respondWithError(res, 500, 'Failed to update loan status');
    }
  }
}

module.exports = new LoanHandler();