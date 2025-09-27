package handler

import (
	"loan-money/internal/auth"
	"loan-money/internal/database"
	"loan-money/internal/handlers"
	"loan-money/pkg/env"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	router        *mux.Router
	isInitialized bool
)

func init() {
	if !isInitialized {
		initializeApp()
		isInitialized = true
	}
}

func initializeApp() {
	// Load environment variables
	if err := env.LoadEnv(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return
	}

	// Create database tables if they don't exist
	if err := database.CreateTables(db); err != nil {
		log.Printf("Failed to create database tables: %v", err)
		return
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	profileHandler := handlers.NewProfileHandler(db)
	loanHandler := handlers.NewLoanHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	// Setup routes
	router = mux.NewRouter()

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

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify your frontend domain
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Wrap router with CORS
	router = c.Handler(router).(*mux.Router)
}

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	if router == nil {
		http.Error(w, "Server not initialized", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
