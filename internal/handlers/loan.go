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

	"github.com/gorilla/mux"
)

// LoanHandler handles loan-related requests
type LoanHandler struct {
	db *sql.DB
}

// NewLoanHandler creates a new LoanHandler instance
func NewLoanHandler(db *sql.DB) *LoanHandler {
	return &LoanHandler{db: db}
}

// GetLoans retrieves loans with pagination
func (h *LoanHandler) GetLoans(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE user_id = $1"
	args := []interface{}{user.ID}
	argIndex := 2

	if status != "" && (status == "active" || status == "completed" || status == "overdue") {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if search != "" {
		whereClause += fmt.Sprintf(" AND borrower_name ILIKE $%d", argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM loans %s", whereClause)
	var total int
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to count loans")
		return
	}

	// Get loans with transaction totals
	query := fmt.Sprintf(`
		SELECT 
			l.id, l.borrower_name, l.amount, l.status, 
			l.loan_date, l.due_date, l.created_at, l.updated_at,
			COALESCE(SUM(t.amount), 0) as total_paid
		FROM loans l
		LEFT JOIN transactions t ON l.id = t.loan_id
		%s
		GROUP BY l.id, l.borrower_name, l.amount, l.status, l.loan_date, l.due_date, l.created_at, l.updated_at
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve loans")
		return
	}
	defer rows.Close()

	var loans []models.LoanResponse
	for rows.Next() {
		var loan models.LoanResponse
		err := rows.Scan(
			&loan.ID, &loan.BorrowerName, &loan.Amount, &loan.Status,
			&loan.LoanDate, &loan.DueDate, &loan.CreatedAt, &loan.UpdatedAt,
			&loan.TotalPaid,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan loan data")
			return
		}

		loan.UserID = user.ID
		loan.RemainingDebt = loan.Amount - loan.TotalPaid
		loans = append(loans, loan)
	}

	// Calculate pagination
	pages := int(math.Ceil(float64(total) / float64(limit)))

	response := models.PaginatedResponse{
		Data: loans,
		Pagination: models.Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: pages,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CreateLoan creates a new loan
func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.BorrowerName == "" {
		respondWithError(w, http.StatusBadRequest, "Borrower name is required")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than 0")
		return
	}

	if req.LoanDate == "" {
		respondWithError(w, http.StatusBadRequest, "Loan date is required")
		return
	}

	// Parse dates
	loanDate, err := time.Parse("2006-01-02", req.LoanDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid loan date format (use YYYY-MM-DD)")
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		dueDateParsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid due date format (use YYYY-MM-DD)")
			return
		}
		dueDate = &dueDateParsed
	}

	// Create loan
	var loan models.Loan
	query := `
		INSERT INTO loans (user_id, borrower_name, amount, loan_date, due_date) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, user_id, borrower_name, amount, status, loan_date, due_date, created_at, updated_at`

	err = h.db.QueryRow(query, user.ID, req.BorrowerName, req.Amount, loanDate, dueDate).Scan(
		&loan.ID, &loan.UserID, &loan.BorrowerName, &loan.Amount, &loan.Status,
		&loan.LoanDate, &loan.DueDate, &loan.CreatedAt, &loan.UpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create loan")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Loan created successfully",
		"loan":    loan,
	})
}

// GetLoan retrieves a specific loan by ID
func (h *LoanHandler) GetLoan(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	loanID := vars["id"]

	if loanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	// Get loan with transaction totals
	var loan models.LoanResponse
	query := `
		SELECT 
			l.id, l.borrower_name, l.amount, l.status, 
			l.loan_date, l.due_date, l.created_at, l.updated_at,
			COALESCE(SUM(t.amount), 0) as total_paid
		FROM loans l
		LEFT JOIN transactions t ON l.id = t.loan_id
		WHERE l.id = $1 AND l.user_id = $2
		GROUP BY l.id, l.borrower_name, l.amount, l.status, l.loan_date, l.due_date, l.created_at, l.updated_at`

	err := h.db.QueryRow(query, loanID, user.ID).Scan(
		&loan.ID, &loan.BorrowerName, &loan.Amount, &loan.Status,
		&loan.LoanDate, &loan.DueDate, &loan.CreatedAt, &loan.UpdatedAt,
		&loan.TotalPaid,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Loan not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve loan")
		return
	}

	loan.UserID = user.ID
	loan.RemainingDebt = loan.Amount - loan.TotalPaid

	respondWithJSON(w, http.StatusOK, loan)
}

// UpdateLoan updates an existing loan
func (h *LoanHandler) UpdateLoan(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	loanID := vars["id"]

	if loanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	var req models.LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.BorrowerName == "" {
		respondWithError(w, http.StatusBadRequest, "Borrower name is required")
		return
	}

	if req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Amount must be greater than 0")
		return
	}

	// Parse dates
	loanDate, err := time.Parse("2006-01-02", req.LoanDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid loan date format (use YYYY-MM-DD)")
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		dueDateParsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid due date format (use YYYY-MM-DD)")
			return
		}
		dueDate = &dueDateParsed
	}

	// Check if loan exists and belongs to user
	var existingLoanID string
	err = h.db.QueryRow("SELECT id FROM loans WHERE id = $1 AND user_id = $2", loanID, user.ID).Scan(&existingLoanID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Loan not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to verify loan ownership")
		return
	}

	// Update loan
	query := `
		UPDATE loans 
		SET borrower_name = $1, amount = $2, loan_date = $3, due_date = $4, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $5 AND user_id = $6`

	_, err = h.db.Exec(query, req.BorrowerName, req.Amount, loanDate, dueDate, loanID, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update loan")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Loan updated successfully",
	})
}

// DeleteLoan deletes a loan
func (h *LoanHandler) DeleteLoan(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	loanID := vars["id"]

	if loanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	// Check if loan exists and belongs to user
	var existingLoanID string
	err := h.db.QueryRow("SELECT id FROM loans WHERE id = $1 AND user_id = $2", loanID, user.ID).Scan(&existingLoanID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Loan not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to verify loan ownership")
		return
	}

	// Begin transaction to delete loan and its transactions
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Delete transactions first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM transactions WHERE loan_id = $1", loanID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete related transactions")
		return
	}

	// Delete loan
	_, err = tx.Exec("DELETE FROM loans WHERE id = $1 AND user_id = $2", loanID, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete loan")
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Loan deleted successfully",
	})
}

// UpdateLoanStatus updates loan status (active/completed)
func (h *LoanHandler) UpdateLoanStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	loanID := vars["id"]

	if loanID == "" {
		respondWithError(w, http.StatusBadRequest, "Loan ID is required")
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Status != "active" && req.Status != "completed" && req.Status != "overdue" {
		respondWithError(w, http.StatusBadRequest, "Status must be either 'active', 'completed', or 'overdue'")
		return
	}

	// Check if loan exists and belongs to user
	var existingLoanID string
	err := h.db.QueryRow("SELECT id FROM loans WHERE id = $1 AND user_id = $2", loanID, user.ID).Scan(&existingLoanID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Loan not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to verify loan ownership")
		return
	}

	// Update loan status
	_, err = h.db.Exec(`
		UPDATE loans 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND user_id = $3`, req.Status, loanID, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update loan status")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Loan status updated successfully",
	})
}
