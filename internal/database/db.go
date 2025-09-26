package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// InitDB initializes database connection
func InitDB() (*sql.DB, error) {
	// Get database configuration from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "loan_money"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	fmt.Printf("Connecting to database: %s:%s/%s (SSL: %s)\n", host, port, dbname, sslmode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for cloud databases
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database")
	return db, nil
}

// CreateTables creates the necessary database tables if they don't exist
func CreateTables(db *sql.DB) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR NOT NULL UNIQUE,
			password VARCHAR NOT NULL,
			full_name VARCHAR,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS loans (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			borrower_name VARCHAR NOT NULL,
			amount NUMERIC NOT NULL,
			status VARCHAR NOT NULL DEFAULT 'active',
			loan_date DATE NOT NULL,
			due_date DATE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			loan_id UUID NOT NULL REFERENCES loans(id),
			amount NUMERIC NOT NULL,
			remark TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	fmt.Println("Database tables created successfully")
	return nil
}
