package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"loan-money/internal/auth"
	"loan-money/internal/database"
	"loan-money/internal/handlers"
	"loan-money/pkg/env"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the request details
		duration := time.Since(start)
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		log.Printf("[%s] %s %s %s - Status: %d - Duration: %v - IP: %s",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			lrw.statusCode,
			duration,
			clientIP,
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func main() {
	// Load environment variables from .env file
	if err := env.LoadEnv(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create database tables if they don't exist
	if err := database.CreateTables(db); err != nil {
		log.Fatal("Failed to create database tables:", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	profileHandler := handlers.NewProfileHandler(db)
	loanHandler := handlers.NewLoanHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	// Setup routes
	router := mux.NewRouter()

	// Add logging middleware to all routes
	router.Use(LoggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public auth routes (no authentication required)
	api.HandleFunc("/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware)

	// Profile management endpoints
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/profile", profileHandler.UpdateProfile).Methods("PATCH")
	protected.HandleFunc("/change-password", profileHandler.ChangePassword).Methods("PATCH")

	// Dashboard endpoints
	protected.HandleFunc("/dashboard/stats", dashboardHandler.GetDashboardStats).Methods("GET")
	protected.HandleFunc("/dashboard/recent-transactions", dashboardHandler.GetRecentTransactions).Methods("GET")
	protected.HandleFunc("/dashboard/loan-summary", dashboardHandler.GetLoanSummary).Methods("GET")
	protected.HandleFunc("/dashboard/monthly-stats", dashboardHandler.GetMonthlyStats).Methods("GET")
	protected.HandleFunc("/dashboard/overdue-loans", dashboardHandler.GetOverdueLoans).Methods("GET")

	// Loan management endpoints
	protected.HandleFunc("/loans", loanHandler.GetLoans).Methods("GET")
	protected.HandleFunc("/loans", loanHandler.CreateLoan).Methods("POST")
	protected.HandleFunc("/loans/{id}", loanHandler.GetLoan).Methods("GET")
	protected.HandleFunc("/loans/{id}", loanHandler.UpdateLoan).Methods("PATCH")
	protected.HandleFunc("/loans/{id}", loanHandler.DeleteLoan).Methods("DELETE")
	protected.HandleFunc("/loans/{id}/status", loanHandler.UpdateLoanStatus).Methods("PATCH")

	// Transaction management endpoints
	protected.HandleFunc("/transactions", transactionHandler.GetTransactions).Methods("GET")
	protected.HandleFunc("/transactions", transactionHandler.CreateTransaction).Methods("POST")
	protected.HandleFunc("/transactions/{id}", transactionHandler.GetTransaction).Methods("GET")
	protected.HandleFunc("/transactions/{id}", transactionHandler.UpdateTransaction).Methods("PATCH")
	protected.HandleFunc("/transactions/{id}", transactionHandler.DeleteTransaction).Methods("DELETE")
	protected.HandleFunc("/loans/{loan_id}/transactions", transactionHandler.GetTransactionsByLoan).Methods("GET")

	// Serve static files (HTML, CSS, JS, images)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify your frontend domain
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Wrap router with CORS
	handler := c.Handler(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
