package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"loan-money/internal/models"
)

// RespondWithError sends error response and logs it
func RespondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("API Error: %d - %s", code, message)
	RespondWithJSON(w, code, models.ErrorResponse{Error: message})
}

// RespondWithJSON sends JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("JSON Encoding Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to encode response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// LogDatabaseError logs database errors with more context
func LogDatabaseError(operation string, err error) {
	log.Printf("Database Error [%s]: %v", operation, err)
}

// LogAPICall logs API calls with details
func LogAPICall(method, path, userID string, statusCode int) {
	log.Printf("API Call: %s %s - User: %s - Status: %d", method, path, userID, statusCode)
}
