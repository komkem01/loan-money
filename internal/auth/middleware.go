package auth

import (
	"context"
	"net/http"
	"strings"

	"loan-money/internal/models"
	"loan-money/pkg/utils"

	"github.com/google/uuid"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
)

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateJWT(tokenParts[1])
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Create user object from claims
		user := models.User{
			ID:       claims.UserID,
			Username: claims.Username,
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(models.User)
	return &user, ok
}

// OptionalAuthMiddleware is a middleware that tries to extract user from token but doesn't require authentication
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateJWT(tokenParts[1])
		if err != nil {
			// Invalid token, continue without user
			next.ServeHTTP(w, r)
			return
		}

		// Create user object from claims
		user := models.User{
			ID:       claims.UserID,
			Username: claims.Username,
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// RequireUserID middleware checks if the user ID in the URL matches the authenticated user
func RequireUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (should be set by AuthMiddleware)
		user, ok := GetUserFromContext(r)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "User not authenticated")
			return
		}

		// Extract user ID from URL path (you might need to adjust this based on your routing)
		// This is a simple example assuming the URL format is /api/v1/users/{userID}/...
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 5 || pathParts[4] == "" {
			respondWithError(w, http.StatusBadRequest, "User ID not found in URL")
			return
		}

		userIDStr := pathParts[4]
		userIDFromURL, err := uuid.Parse(userIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		// Check if the user ID matches
		if user.ID != userIDFromURL {
			respondWithError(w, http.StatusForbidden, "Access denied: user ID mismatch")
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
