package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection string from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/panda_pocket?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Read SQL migration file
	sqlFile := filepath.Join("scripts", "add_user_role.sql")
	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	// Execute migration
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	fmt.Println("Migration completed successfully!")
	fmt.Println("Added role column to users table with default value 'user'")
	fmt.Println("Added check constraint for valid roles: user, admin, super_admin")
	fmt.Println("Created index on role column for better performance")
}
