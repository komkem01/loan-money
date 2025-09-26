package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"loan-money/internal/auth"
	"loan-money/internal/database"
	"loan-money/internal/handlers"
	"loan-money/pkg/env"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

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

	// Setup routes
	router := mux.NewRouter()

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
	// fmt.Println("Available endpoints:")
	// fmt.Println("GET  /health - Health check")
	// fmt.Println("POST /api/v1/register - User registration")
	// fmt.Println("POST /api/v1/login - User login")
	// fmt.Println("GET  /api/v1/profile - Get user profile (requires auth)")
	// fmt.Println("Press Ctrl+C to stop the server")

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
