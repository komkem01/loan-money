package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // Hidden from JSON output
	FullName  *string   `json:"full_name,omitempty" db:"full_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
}

// Transaction represents a transaction in the system
type Transaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	LoanID    uuid.UUID `json:"loan_id" db:"loan_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Remark    *string   `json:"remark,omitempty" db:"remark"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
