package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"loan-money/internal/auth"
	"loan-money/internal/models"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	db *sql.DB
}

// NewDashboardHandler creates a new DashboardHandler instance
func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetDashboardStats retrieves dashboard statistics
func (h *DashboardHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var stats models.DashboardStats

	// Get loan counts and amounts
	err := h.db.QueryRow(`
		SELECT 
			COUNT(*) as total_loans,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_loans,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_loans,
			COALESCE(SUM(amount), 0) as total_loan_amount
		FROM loans 
		WHERE user_id = $1
	`, user.ID).Scan(
		&stats.TotalLoans,
		&stats.ActiveLoans,
		&stats.CompletedLoans,
		&stats.TotalLoanAmount,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve loan statistics")
		return
	}

	// Get total paid amount
	err = h.db.QueryRow(`
		SELECT COALESCE(SUM(t.amount), 0) as total_paid
		FROM transactions t
		JOIN loans l ON t.loan_id = l.id
		WHERE l.user_id = $1
	`, user.ID).Scan(&stats.TotalPaidAmount)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve transaction statistics")
		return
	}

	// Calculate total debt amount
	stats.TotalDebtAmount = stats.TotalLoanAmount - stats.TotalPaidAmount

	respondWithJSON(w, http.StatusOK, stats)
}

// GetRecentTransactions retrieves recent transactions for dashboard
func (h *DashboardHandler) GetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse limit parameter (default: 5, max: 20)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 20 {
		limit = 5
	}

	query := `
		SELECT 
			t.id, t.loan_id, t.amount, t.remark, t.created_at,
			l.borrower_name, l.amount as loan_amount, l.status as loan_status
		FROM transactions t
		JOIN loans l ON t.loan_id = l.id
		WHERE l.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2`

	rows, err := h.db.Query(query, user.ID, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve recent transactions")
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var transaction models.Transaction
		var borrowerName string
		var loanAmount float64
		var loanStatus string

		err := rows.Scan(
			&transaction.ID, &transaction.LoanID, &transaction.Amount,
			&transaction.Remark, &transaction.CreatedAt,
			&borrowerName, &loanAmount, &loanStatus,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan transaction data")
			return
		}

		transactionWithLoan := map[string]interface{}{
			"id":            transaction.ID,
			"loan_id":       transaction.LoanID,
			"amount":        transaction.Amount,
			"remark":        transaction.Remark,
			"created_at":    transaction.CreatedAt,
			"borrower_name": borrowerName,
			"loan_amount":   loanAmount,
			"loan_status":   loanStatus,
		}

		transactions = append(transactions, transactionWithLoan)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// GetLoanSummary retrieves loan summary with payment status
func (h *DashboardHandler) GetLoanSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse limit parameter (default: 10, max: 50)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	status := r.URL.Query().Get("status")

	// Build WHERE clause
	whereClause := "WHERE l.user_id = $1"
	args := []interface{}{user.ID}
	argIndex := 2

	if status != "" && (status == "active" || status == "completed") {
		whereClause += " AND l.status = $2"
		args = append(args, status)
		argIndex++
	}

	query := `
		SELECT 
			l.id, l.borrower_name, l.amount, l.status, 
			l.loan_date, l.due_date, l.created_at, l.updated_at,
			COALESCE(SUM(t.amount), 0) as total_paid,
			(l.amount - COALESCE(SUM(t.amount), 0)) as remaining_debt
		FROM loans l
		LEFT JOIN transactions t ON l.id = t.loan_id
		` + whereClause + `
		GROUP BY l.id, l.borrower_name, l.amount, l.status, l.loan_date, l.due_date, l.created_at, l.updated_at
		ORDER BY l.created_at DESC
		LIMIT $` + strconv.Itoa(argIndex)

	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve loan summary")
		return
	}
	defer rows.Close()

	var loans []models.LoanResponse
	for rows.Next() {
		var loan models.LoanResponse
		err := rows.Scan(
			&loan.ID, &loan.BorrowerName, &loan.Amount, &loan.Status,
			&loan.LoanDate, &loan.DueDate, &loan.CreatedAt, &loan.UpdatedAt,
			&loan.TotalPaid, &loan.RemainingDebt,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan loan data")
			return
		}

		loan.UserID = user.ID
		loans = append(loans, loan)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"loans": loans,
		"count": len(loans),
	})
}

// GetMonthlyStats retrieves monthly statistics for charts
func (h *DashboardHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get monthly loan amounts for the last 12 months
	loanStatsQuery := `
		SELECT 
			TO_CHAR(loan_date, 'YYYY-MM') as month,
			COUNT(*) as loan_count,
			SUM(amount) as total_amount
		FROM loans 
		WHERE user_id = $1 
		AND loan_date >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY TO_CHAR(loan_date, 'YYYY-MM')
		ORDER BY month DESC`

	rows, err := h.db.Query(loanStatsQuery, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve monthly loan statistics")
		return
	}
	defer rows.Close()

	var loanStats []map[string]interface{}
	for rows.Next() {
		var month string
		var count int
		var amount float64

		err := rows.Scan(&month, &count, &amount)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan loan statistics")
			return
		}

		loanStats = append(loanStats, map[string]interface{}{
			"month":  month,
			"count":  count,
			"amount": amount,
		})
	}

	// Get monthly payment amounts for the last 12 months
	paymentStatsQuery := `
		SELECT 
			TO_CHAR(t.created_at, 'YYYY-MM') as month,
			COUNT(*) as payment_count,
			SUM(t.amount) as total_amount
		FROM transactions t
		JOIN loans l ON t.loan_id = l.id
		WHERE l.user_id = $1 
		AND t.created_at >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY TO_CHAR(t.created_at, 'YYYY-MM')
		ORDER BY month DESC`

	rows2, err := h.db.Query(paymentStatsQuery, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve monthly payment statistics")
		return
	}
	defer rows2.Close()

	var paymentStats []map[string]interface{}
	for rows2.Next() {
		var month string
		var count int
		var amount float64

		err := rows2.Scan(&month, &count, &amount)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan payment statistics")
			return
		}

		paymentStats = append(paymentStats, map[string]interface{}{
			"month":  month,
			"count":  count,
			"amount": amount,
		})
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"loan_stats":    loanStats,
		"payment_stats": paymentStats,
	})
}

// GetOverdueLoans retrieves loans that are overdue
func (h *DashboardHandler) GetOverdueLoans(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := `
		SELECT 
			l.id, l.borrower_name, l.amount, l.status, 
			l.loan_date, l.due_date, l.created_at, l.updated_at,
			COALESCE(SUM(t.amount), 0) as total_paid,
			(l.amount - COALESCE(SUM(t.amount), 0)) as remaining_debt,
			(CURRENT_DATE - l.due_date) as days_overdue
		FROM loans l
		LEFT JOIN transactions t ON l.id = t.loan_id
		WHERE l.user_id = $1 
		AND l.status = 'active'
		AND l.due_date < CURRENT_DATE
		AND (l.amount - COALESCE(SUM(t.amount), 0)) > 0
		GROUP BY l.id, l.borrower_name, l.amount, l.status, l.loan_date, l.due_date, l.created_at, l.updated_at
		ORDER BY l.due_date ASC`

	rows, err := h.db.Query(query, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve overdue loans")
		return
	}
	defer rows.Close()

	var overdueLoans []map[string]interface{}
	for rows.Next() {
		var loan models.LoanResponse
		var daysOverdue int

		err := rows.Scan(
			&loan.ID, &loan.BorrowerName, &loan.Amount, &loan.Status,
			&loan.LoanDate, &loan.DueDate, &loan.CreatedAt, &loan.UpdatedAt,
			&loan.TotalPaid, &loan.RemainingDebt, &daysOverdue,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan overdue loan data")
			return
		}

		loan.UserID = user.ID

		overdueLoans = append(overdueLoans, map[string]interface{}{
			"loan":         loan,
			"days_overdue": daysOverdue,
		})
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"overdue_loans": overdueLoans,
		"count":         len(overdueLoans),
	})
}
