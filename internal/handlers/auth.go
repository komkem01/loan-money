package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"loan-money/internal/models"
	"loan-money/pkg/utils"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	if len(req.Username) < 3 {
		respondWithError(w, http.StatusBadRequest, "Username must be at least 3 characters long")
		return
	}

	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters long")
		return
	}

	// Check if user already exists
	var existingUserID string
	err := h.db.QueryRow("SELECT id FROM users WHERE username = $1", req.Username).Scan(&existingUserID)
	if err != sql.ErrNoRows {
		if err == nil {
			respondWithError(w, http.StatusConflict, "Username already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	var user models.User
	query := `
		INSERT INTO users (username, password, full_name) 
		VALUES ($1, $2, $3) 
		RETURNING id, username, full_name, created_at`

	err = h.db.QueryRow(query, req.Username, hashedPassword, req.FullName).Scan(
		&user.ID, &user.Username, &user.FullName, &user.CreatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user from database
	var user models.User
	var hashedPassword string
	query := "SELECT id, username, password, full_name, created_at FROM users WHERE username = $1"

	err := h.db.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.FullName, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Verify password
	isValid, err := utils.VerifyPassword(req.Password, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password verification error")
		return
	}

	if !isValid {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetUserFromToken extracts user information from JWT token
func (h *AuthHandler) GetUserFromToken(r *http.Request) (*models.User, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	// Extract token from "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	// Validate JWT token
	claims, err := utils.ValidateJWT(tokenParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user from database
	var user models.User
	query := "SELECT id, username, full_name, created_at FROM users WHERE id = $1"

	err = h.db.QueryRow(query, claims.UserID).Scan(
		&user.ID, &user.Username, &user.FullName, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to encode response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
