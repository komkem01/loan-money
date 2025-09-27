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
			password_hash VARCHAR NOT NULL,
			full_name VARCHAR,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// Migration queries to handle existing data
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();`,

		// Rename password column to password_hash if it exists
		`DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'password') THEN
				ALTER TABLE users RENAME COLUMN password TO password_hash;
			END IF;
		END $$;`,

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
			created_at TIMESTAMPTZ DEFAULT NOW(),
			payment_date TIMESTAMP,
			deleted_at TIMESTAMP,
			updated_at TIMESTAMP
		);`,

		// Add missing columns to transactions table if they don't exist
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_date TIMESTAMP;`,
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;`,
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP;`,

		// Add missing columns to loans table if they don't exist
		`ALTER TABLE loans ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	fmt.Println("Database tables created successfully")
	return nil
}
