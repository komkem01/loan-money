package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"loan-money/internal/auth"
	"loan-money/internal/models"
	"loan-money/pkg/utils"
)

// ProfileHandler handles profile-related requests
type ProfileHandler struct {
	db *sql.DB
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// GetProfile retrieves user profile information
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Query user data from database
	var userProfile models.User
	err := h.db.QueryRow(`
		SELECT id, username, full_name, created_at 
		FROM users 
		WHERE id = $1
	`, user.ID).Scan(&userProfile.ID, &userProfile.Username, &userProfile.FullName, &userProfile.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}

	// Create response without sensitive data
	profileResponse := map[string]interface{}{
		"id":         userProfile.ID,
		"username":   userProfile.Username,
		"full_name":  userProfile.FullName,
		"created_at": userProfile.CreatedAt,
	}

	respondWithJSON(w, http.StatusOK, profileResponse)
}

// UpdateProfileRequest represents the request body for updating profile
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

// UpdateProfile updates user profile information (PATCH method)
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.FullName == "" {
		respondWithError(w, http.StatusBadRequest, "Full name is required")
		return
	}

	if len(strings.TrimSpace(req.FullName)) < 2 {
		respondWithError(w, http.StatusBadRequest, "Full name must be at least 2 characters long")
		return
	}

	// Update user profile in database
	_, err := h.db.Exec(`
		UPDATE users 
		SET full_name = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, strings.TrimSpace(req.FullName), user.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Return updated profile data
	var updatedProfile models.User
	err = h.db.QueryRow(`
		SELECT id, username, full_name, created_at 
		FROM users 
		WHERE id = $1
	`, user.ID).Scan(&updatedProfile.ID, &updatedProfile.Username, &updatedProfile.FullName, &updatedProfile.CreatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated profile")
		return
	}

	profileResponse := map[string]interface{}{
		"id":         updatedProfile.ID,
		"username":   updatedProfile.Username,
		"full_name":  updatedProfile.FullName,
		"created_at": updatedProfile.CreatedAt,
		"message":    "Profile updated successfully",
	}

	respondWithJSON(w, http.StatusOK, profileResponse)
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword updates user password (PATCH method)
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.CurrentPassword == "" || req.NewPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	if len(req.NewPassword) < 6 {
		respondWithError(w, http.StatusBadRequest, "New password must be at least 6 characters long")
		return
	}

	if req.CurrentPassword == req.NewPassword {
		respondWithError(w, http.StatusBadRequest, "New password must be different from current password")
		return
	}

	// Get current password hash from database
	var currentPasswordHash string
	err := h.db.QueryRow("SELECT password_hash FROM users WHERE id = $1", user.ID).Scan(&currentPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to verify current password")
		return
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, currentPasswordHash) {
		respondWithError(w, http.StatusBadRequest, "Current password is incorrect")
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process new password")
		return
	}

	// Update password in database
	_, err = h.db.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, newPasswordHash, user.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Password changed successfully",
	}

	respondWithJSON(w, http.StatusOK, response)
}
