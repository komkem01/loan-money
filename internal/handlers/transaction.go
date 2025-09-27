package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"loan-money/internal/auth"
	"loan-money/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	db *sql.DB
}

// NewTransactionHandler creates a new TransactionHandler instance
func NewTransactionHandler(db *sql.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

// GetTransactions retrieves transactions with pagination
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	loanID := r.URL.Query().Get("loan_id")
	search := r.URL.Query().Get("search")

	offset := (page - 1) * limit

	// Build query based on filters
	var query string
	var countQuery string
	var args []any
	var countArgs []any

	baseQuery := `
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE l.user_id = $1 AND t.deleted_at IS NULL
	`

	baseCountQuery := `
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE l.user_id = $1 AND t.deleted_at IS NULL
	`

	args = append(args, user.ID)
	countArgs = append(countArgs, user.ID)
	paramCount := 1

	// Add loan_id filter if provided
	if loanID != "" {
		paramCount++
		baseQuery += fmt.Sprintf(" AND t.loan_id = $%d", paramCount)
		baseCountQuery += fmt.Sprintf(" AND t.loan_id = $%d", paramCount)
		args = append(args, loanID)
		countArgs = append(countArgs, loanID)
	}

	// Add search filter if provided
	if search != "" {
		paramCount++
		baseQuery += fmt.Sprintf(" AND (l.borrower_name ILIKE $%d OR t.remark ILIKE $%d)", paramCount, paramCount)
		baseCountQuery += fmt.Sprintf(" AND (l.borrower_name ILIKE $%d OR t.remark ILIKE $%d)", paramCount, paramCount)
		searchParam := "%" + search + "%"
		args = append(args, searchParam)
		countArgs = append(countArgs, searchParam)
	}

	// Complete queries - Match actual database schema
	query = `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			t.payment_date, t.deleted_at, t.updated_at,
			l.borrower_name, l.amount as loan_amount
	` + baseQuery + `
		ORDER BY t.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", paramCount+1) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+2)

	countQuery = `SELECT COUNT(*) ` + baseCountQuery

	args = append(args, limit, offset)

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count transactions")
		return
	}

	// Get transactions
	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}
	defer rows.Close()

	var transactions []map[string]any
	for rows.Next() {
		var t models.Transaction
		var borrowerName string
		var loanAmount float64

		var paymentDate sql.NullTime
		var deletedAt sql.NullTime
		var updatedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.LoanID, &t.Amount, &t.Remark, &t.CreatedAt,
			&paymentDate, &deletedAt, &updatedAt,
			&borrowerName, &loanAmount,
		)

		if paymentDate.Valid {
			dateStr := paymentDate.Time.Format("2006-01-02")
			t.PaymentDate = &dateStr
		}
		if deletedAt.Valid {
			deletedAtStr := deletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			t.DeletedAt = &deletedAtStr
		}
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan transaction")
			return
		}

		transactionData := map[string]any{
			"id":            t.ID,
			"loan_id":       t.LoanID,
			"amount":        t.Amount,
			"remark":        t.Remark,
			"payment_date":  t.PaymentDate,
			"created_at":    t.CreatedAt,
			"updated_at":    t.UpdatedAt,
			"borrower_name": borrowerName,
			"loan_amount":   loanAmount,
		}

		transactions = append(transactions, transactionData)
	}

	// Calculate pagination
	pages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedResponse{
		Data: transactions,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: pages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetTransaction retrieves a specific transaction by ID
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	transactionID := vars["id"]

	query := `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			t.payment_date, t.deleted_at, t.updated_at,
			l.borrower_name, l.amount as loan_amount
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE t.id = $1 AND l.user_id = $2 AND t.deleted_at IS NULL
	`

	var t models.Transaction
	var borrowerName string
	var loanAmount float64

	err := h.db.QueryRow(query, transactionID, user.ID).Scan(
		&t.ID, &t.LoanID, &t.Amount, &t.Remark, &t.CreatedAt,
		&t.PaymentDate, &t.DeletedAt, &t.UpdatedAt,
		&borrowerName, &loanAmount,
	)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Transaction not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch transaction")
		return
	}

	response := map[string]any{
		"id":            t.ID,
		"loan_id":       t.LoanID,
		"amount":        t.Amount,
		"remark":        t.Remark,
		"payment_date":  t.PaymentDate,
		"created_at":    t.CreatedAt,
		"updated_at":    t.UpdatedAt,
		"borrower_name": borrowerName,
		"loan_amount":   loanAmount,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CreateTransaction creates a new transaction
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.LoanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than 0")
		return
	}

	// Get loan details and verify ownership
	var loanAmountCheck float64
	var currentTotalPaid float64
	err := h.db.QueryRow(`
		SELECT 
			l.amount,
			COALESCE(SUM(t.amount), 0) as total_paid
		FROM loans l
		LEFT JOIN transactions t ON l.id = t.loan_id AND t.deleted_at IS NULL
		WHERE l.id = $1 AND l.user_id = $2 AND l.deleted_at IS NULL
		GROUP BY l.id, l.amount
	`, req.LoanID, user.ID).Scan(&loanAmountCheck, &currentTotalPaid)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusBadRequest, "Loan not found or access denied")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to verify loan")
		return
	}

	// Check if payment amount exceeds remaining debt
	remainingDebt := loanAmountCheck - currentTotalPaid
	if remainingDebt <= 0 {
		respondWithError(w, http.StatusBadRequest, "This loan is already fully paid")
		return
	}
	if req.Amount > remainingDebt {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Payment amount (฿%.2f) exceeds remaining debt (฿%.2f)", req.Amount, remainingDebt))
		return
	}

	// Parse payment date if provided
	var paymentDate *time.Time
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			// Try with datetime format
			parsed, err = time.Parse("2006-01-02T15:04:05", *req.PaymentDate)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Invalid payment date format. Use YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS")
				return
			}
		}
		paymentDate = &parsed
	}

	// Create transaction in a transaction (database transaction)
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	transactionID := uuid.New()
	now := time.Now()

	// Insert the payment transaction
	query := `
		INSERT INTO transactions (id, loan_id, amount, remark, payment_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(query, transactionID, req.LoanID, req.Amount, req.Remark, paymentDate, now, now)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	// Check if loan is fully paid after this payment
	newTotalPaid := currentTotalPaid + req.Amount
	if newTotalPaid >= loanAmountCheck {
		// Update loan status to 'completed' when fully paid
		updateLoanQuery := `
			UPDATE loans 
			SET status = 'completed', updated_at = $1 
			WHERE id = $2
		`
		_, err = tx.Exec(updateLoanQuery, now, req.LoanID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update loan status")
			return
		}
	}

	// Commit the database transaction
	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Fetch the created transaction with loan details
	var transaction models.Transaction
	var borrowerName string
	var loanAmount float64

	fetchQuery := `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			t.payment_date, t.deleted_at, t.updated_at,
			l.borrower_name, l.amount as loan_amount
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE t.id = $1
	`

	err = h.db.QueryRow(fetchQuery, transactionID).Scan(
		&transaction.ID, &transaction.LoanID, &transaction.Amount, &transaction.Remark, &transaction.CreatedAt,
		&transaction.PaymentDate, &transaction.DeletedAt, &transaction.UpdatedAt,
		&borrowerName, &loanAmount,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch created transaction")
		return
	}

	response := map[string]any{
		"id":            transaction.ID,
		"loan_id":       transaction.LoanID,
		"amount":        transaction.Amount,
		"remark":        transaction.Remark,
		"payment_date":  transaction.PaymentDate,
		"created_at":    transaction.CreatedAt,
		"updated_at":    transaction.UpdatedAt,
		"borrower_name": borrowerName,
		"loan_amount":   loanAmount,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// UpdateTransaction updates an existing transaction
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	transactionID := vars["id"]

	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.LoanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than 0")
		return
	}

	// Verify transaction belongs to user
	var transactionExists bool
	err := h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM transactions t
			INNER JOIN loans l ON t.loan_id = l.id
			WHERE t.id = $1 AND l.user_id = $2 AND t.deleted_at IS NULL
		)`,
		transactionID, user.ID,
	).Scan(&transactionExists)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to verify transaction")
		return
	}

	if !transactionExists {
		respondWithError(w, http.StatusNotFound, "Transaction not found or access denied")
		return
	}

	// Get current transaction amount and loan details
	var currentTransactionAmount float64
	var loanAmountUpdate float64
	var currentTotalPaidUpdate float64
	err = h.db.QueryRow(`
		SELECT 
			t.amount,
			l.amount as loan_amount,
			COALESCE((SELECT SUM(t2.amount) FROM transactions t2 WHERE t2.loan_id = l.id AND t2.deleted_at IS NULL AND t2.id != t.id), 0) as total_paid_excluding_this
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE t.id = $1 AND l.user_id = $2 AND t.deleted_at IS NULL
	`, transactionID, user.ID).Scan(&currentTransactionAmount, &loanAmountUpdate, &currentTotalPaidUpdate)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Transaction not found or access denied")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get transaction details")
		return
	}

	// Check if new payment amount exceeds remaining debt (excluding current transaction)
	remainingDebtUpdate := loanAmountUpdate - currentTotalPaidUpdate
	if remainingDebtUpdate <= 0 {
		respondWithError(w, http.StatusBadRequest, "This loan is already fully paid")
		return
	}
	if req.Amount > remainingDebtUpdate {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Payment amount (฿%.2f) exceeds remaining debt (฿%.2f)", req.Amount, remainingDebtUpdate))
		return
	}

	// Parse payment date if provided
	var paymentDate *time.Time
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			// Try with datetime format
			parsed, err = time.Parse("2006-01-02T15:04:05", *req.PaymentDate)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Invalid payment date format. Use YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS")
				return
			}
		}
		paymentDate = &parsed
	}

	// Update transaction in a database transaction
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	now := time.Now()
	query := `
		UPDATE transactions 
		SET loan_id = $1, amount = $2, remark = $3, payment_date = $4, updated_at = $5
		WHERE id = $6
	`

	_, err = tx.Exec(query, req.LoanID, req.Amount, req.Remark, paymentDate, now, transactionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update transaction")
		return
	}

	// Check if loan is fully paid after this update
	newTotalPaidUpdate := currentTotalPaidUpdate + req.Amount
	if newTotalPaidUpdate >= loanAmountUpdate {
		// Update loan status to 'completed'
		updateLoanQuery := `
			UPDATE loans 
			SET status = 'completed', updated_at = $1 
			WHERE id = $2
		`
		_, err = tx.Exec(updateLoanQuery, now, req.LoanID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update loan status")
			return
		}
	} else {
		// If the loan was previously completed but now isn't (due to amount reduction), revert to active
		checkAndRevertQuery := `
			UPDATE loans 
			SET status = 'active', updated_at = $1 
			WHERE id = $2 AND status = 'completed'
		`
		_, err = tx.Exec(checkAndRevertQuery, now, req.LoanID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to revert loan status")
			return
		}
	}

	// Commit the database transaction
	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Fetch the updated transaction with loan details
	var transaction models.Transaction
	var borrowerName string
	var loanAmount float64

	fetchQuery := `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			t.payment_date, t.deleted_at, t.updated_at,
			l.borrower_name, l.amount as loan_amount
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE t.id = $1
	`

	err = h.db.QueryRow(fetchQuery, transactionID).Scan(
		&transaction.ID, &transaction.LoanID, &transaction.Amount, &transaction.Remark, &transaction.CreatedAt,
		&transaction.PaymentDate, &transaction.DeletedAt, &transaction.UpdatedAt,
		&borrowerName, &loanAmount,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch updated transaction")
		return
	}

	response := map[string]any{
		"id":            transaction.ID,
		"loan_id":       transaction.LoanID,
		"amount":        transaction.Amount,
		"remark":        transaction.Remark,
		"payment_date":  transaction.PaymentDate,
		"created_at":    transaction.CreatedAt,
		"updated_at":    transaction.UpdatedAt,
		"borrower_name": borrowerName,
		"loan_amount":   loanAmount,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// DeleteTransaction soft deletes a transaction
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	transactionID := vars["id"]

	// Get transaction details including loan info
	var loanID string
	var loanAmountDelete float64
	var transactionAmount float64
	var currentTotalPaidDelete float64
	err := h.db.QueryRow(`
		SELECT 
			t.loan_id,
			t.amount,
			l.amount as loan_amount,
			COALESCE((SELECT SUM(t2.amount) FROM transactions t2 WHERE t2.loan_id = l.id AND t2.deleted_at IS NULL), 0) as total_paid
		FROM transactions t
		INNER JOIN loans l ON t.loan_id = l.id
		WHERE t.id = $1 AND l.user_id = $2 AND t.deleted_at IS NULL
	`, transactionID, user.ID).Scan(&loanID, &transactionAmount, &loanAmountDelete, &currentTotalPaidDelete)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Transaction not found or access denied")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get transaction details")
		return
	}

	// Delete transaction in a database transaction
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Soft delete transaction
	now := time.Now()
	query := `UPDATE transactions SET deleted_at = $1 WHERE id = $2`

	_, err = tx.Exec(query, now, transactionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete transaction")
		return
	}

	// Calculate new total paid after deleting this transaction
	newTotalPaidDelete := currentTotalPaidDelete - transactionAmount

	// Update loan status based on new total paid
	if newTotalPaidDelete >= loanAmountDelete {
		// Loan is still fully paid
		updateLoanQuery := `
			UPDATE loans 
			SET status = 'completed', updated_at = $1 
			WHERE id = $2
		`
		_, err = tx.Exec(updateLoanQuery, now, loanID)
	} else {
		// Loan is no longer fully paid, revert to active
		updateLoanQuery := `
			UPDATE loans 
			SET status = 'active', updated_at = $1 
			WHERE id = $2 AND status = 'completed'
		`
		_, err = tx.Exec(updateLoanQuery, now, loanID)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update loan status")
		return
	}

	// Commit the database transaction
	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Transaction deleted successfully"})
}

// GetTransactionsByLoan retrieves all transactions for a specific loan
func (h *TransactionHandler) GetTransactionsByLoan(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	loanID := vars["loan_id"]

	// Verify loan belongs to user
	var loanExists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM loans WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL)",
		loanID, user.ID,
	).Scan(&loanExists)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to verify loan")
		return
	}

	if !loanExists {
		respondWithError(w, http.StatusNotFound, "Loan not found or access denied")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM transactions WHERE loan_id = $1 AND deleted_at IS NULL`
	err = h.db.QueryRow(countQuery, loanID).Scan(&total)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count transactions")
		return
	}

	// Get transactions
	query := `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			t.payment_date, t.deleted_at, t.updated_at
		FROM transactions t
		WHERE t.loan_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.Query(query, loanID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID, &t.LoanID, &t.Amount, &t.Remark, &t.CreatedAt,
			&t.PaymentDate, &t.DeletedAt, &t.UpdatedAt,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan transaction")
			return
		}
		transactions = append(transactions, t)
	}

	// Calculate pagination
	pages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedResponse{
		Data: transactions,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: pages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}
