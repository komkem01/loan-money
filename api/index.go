package handler

import (
	"fmt"
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

	// User profile endpoint (example of protected route)
	protected.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user_id":"%s","username":"%s"}`, user.ID, user.Username)
	}).Methods("GET")

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
