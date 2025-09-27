package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"` // Hidden from JSON output
	FullName     *string    `json:"full_name,omitempty" db:"full_name"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Loan represents a loan in the system
type Loan struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	BorrowerName string     `json:"borrower_name" db:"borrower_name"`
	Amount       float64    `json:"amount" db:"amount"`
	Status       string     `json:"status" db:"status"`
	LoanDate     time.Time  `json:"loan_date" db:"loan_date"`
	DueDate      *time.Time `json:"due_date,omitempty" db:"due_date"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Transaction represents a transaction in the system
type Transaction struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	LoanID      uuid.UUID  `json:"loan_id" db:"loan_id"`
	Amount      float64    `json:"amount" db:"amount"`
	Remark      *string    `json:"remark,omitempty" db:"remark"`
	PaymentDate *time.Time `json:"payment_date,omitempty" db:"payment_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// AuthRequest represents login/register request
type AuthRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Password string  `json:"password" validate:"required,min=6"`
	FullName *string `json:"full_name,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// LoanRequest represents create/update loan request
type LoanRequest struct {
	BorrowerName string  `json:"borrower_name" validate:"required,min=2,max=100"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	LoanDate     string  `json:"loan_date" validate:"required"`
	DueDate      *string `json:"due_date,omitempty"`
}

// LoanResponse represents loan response with additional fields
type LoanResponse struct {
	Loan
	TotalPaid     float64 `json:"total_paid"`
	RemainingDebt float64 `json:"remaining_debt"`
}

// TransactionRequest represents create/update transaction request
type TransactionRequest struct {
	LoanID      string  `json:"loan_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Remark      *string `json:"remark,omitempty"`
	PaymentDate *string `json:"payment_date,omitempty"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalLoans      int     `json:"total_loans"`
	ActiveLoans     int     `json:"active_loans"`
	CompletedLoans  int     `json:"completed_loans"`
	TotalLoanAmount float64 `json:"total_loan_amount"`
	TotalPaidAmount float64 `json:"total_paid_amount"`
	TotalDebtAmount float64 `json:"total_debt_amount"`
}
