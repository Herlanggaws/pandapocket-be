package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go <sqlite_file>")
		fmt.Println("Example: go run cmd/migrate/main.go ./panda_pocket.db")
		os.Exit(1)
	}

	sqliteFile := os.Args[1]

	// Connect to SQLite database
	sqliteDB, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatal("Error opening SQLite database:", err)
	}
	defer sqliteDB.Close()

	// Connect to PostgreSQL database
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "panda_pocket")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	postgresDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening PostgreSQL database:", err)
	}
	defer postgresDB.Close()

	// Test PostgreSQL connection
	err = postgresDB.Ping()
	if err != nil {
		log.Fatal("Error connecting to PostgreSQL:", err)
	}

	fmt.Println("Starting migration from SQLite to PostgreSQL...")

	// Migrate users
	fmt.Println("Migrating users...")
	err = migrateUsers(sqliteDB, postgresDB)
	if err != nil {
		log.Fatal("Error migrating users:", err)
	}

	// Migrate categories
	fmt.Println("Migrating categories...")
	err = migrateCategories(sqliteDB, postgresDB)
	if err != nil {
		log.Fatal("Error migrating categories:", err)
	}

	// Migrate expenses
	fmt.Println("Migrating expenses...")
	err = migrateExpenses(sqliteDB, postgresDB)
	if err != nil {
		log.Fatal("Error migrating expenses:", err)
	}

	// Migrate incomes
	fmt.Println("Migrating incomes...")
	err = migrateIncomes(sqliteDB, postgresDB)
	if err != nil {
		log.Fatal("Error migrating incomes:", err)
	}

	fmt.Println("Migration completed successfully!")
}

func migrateUsers(sqliteDB, postgresDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, email, password_hash, created_at FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var email, passwordHash, createdAt string
		err := rows.Scan(&id, &email, &passwordHash, &createdAt)
		if err != nil {
			return err
		}

		_, err = postgresDB.Exec(`
			INSERT INTO users (id, email, password_hash, created_at) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, id, email, passwordHash, createdAt)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migrateCategories(sqliteDB, postgresDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, user_id, name, color, is_default, category_type, created_at FROM categories")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID int
		var name, color, categoryType, createdAt string
		var isDefault bool
		err := rows.Scan(&id, &userID, &name, &color, &isDefault, &categoryType, &createdAt)
		if err != nil {
			return err
		}

		_, err = postgresDB.Exec(`
			INSERT INTO categories (id, user_id, name, color, is_default, category_type, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, id, userID, name, color, isDefault, categoryType, createdAt)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migrateExpenses(sqliteDB, postgresDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, user_id, category_id, amount, description, date, created_at FROM expenses")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, categoryID int
		var amount float64
		var description, date, createdAt string
		err := rows.Scan(&id, &userID, &categoryID, &amount, &description, &date, &createdAt)
		if err != nil {
			return err
		}

		_, err = postgresDB.Exec(`
			INSERT INTO expenses (id, user_id, category_id, amount, description, date, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, id, userID, categoryID, amount, description, date, createdAt)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migrateIncomes(sqliteDB, postgresDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, user_id, category_id, amount, description, date, created_at FROM incomes")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, categoryID int
		var amount float64
		var description, date, createdAt string
		err := rows.Scan(&id, &userID, &categoryID, &amount, &description, &date, &createdAt)
		if err != nil {
			return err
		}

		_, err = postgresDB.Exec(`
			INSERT INTO incomes (id, user_id, category_id, amount, description, date, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, id, userID, categoryID, amount, description, date, createdAt)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
