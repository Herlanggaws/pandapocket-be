package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() (*sql.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite" // default to sqlite for backward compatibility
	}

	var db *sql.DB
	var err error

	switch dbType {
	case "postgres":
		db, err = initPostgresDB()
	case "sqlite":
		db, err = initSQLiteDB()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, err
	}

	// Create tables
	err = createTables(db, dbType)
	if err != nil {
		return nil, err
	}

	log.Printf("Database initialized successfully with %s", dbType)
	return db, nil
}

func initPostgresDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "panda_pocket")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return sql.Open("postgres", psqlInfo)
}

func initSQLiteDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "./panda_pocket.db")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTables(db *sql.DB, dbType string) error {
	var createTablesSQL string

	if dbType == "postgres" {
		createTablesSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS currencies (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			code VARCHAR(3) NOT NULL,
			name VARCHAR(100) NOT NULL,
			symbol VARCHAR(10) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);

		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			name VARCHAR(255) NOT NULL,
			color VARCHAR(7) DEFAULT '#3B82F6',
			is_default BOOLEAN DEFAULT FALSE,
			category_type VARCHAR(20) DEFAULT 'expense',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);

		CREATE TABLE IF NOT EXISTS expenses (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS incomes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS budgets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			period VARCHAR(20) NOT NULL CHECK (period IN ('weekly', 'monthly', 'yearly')),
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id)
		);

		CREATE TABLE IF NOT EXISTS recurring_transactions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
			next_due_date DATE NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS user_preferences (
			id SERIAL PRIMARY KEY,
			user_id INTEGER UNIQUE NOT NULL,
			primary_currency_id INTEGER NOT NULL,
			email_notifications BOOLEAN DEFAULT TRUE,
			budget_alerts BOOLEAN DEFAULT TRUE,
			recurring_reminders BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (primary_currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			type VARCHAR(50) NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);
		`
	} else {
		// SQLite schema (original)
		createTablesSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS currencies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			symbol TEXT NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);

		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#3B82F6',
			is_default BOOLEAN DEFAULT FALSE,
			category_type TEXT DEFAULT 'expense',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);

		CREATE TABLE IF NOT EXISTS expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			date DATE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS incomes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			date DATE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS budgets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			period TEXT NOT NULL CHECK (period IN ('weekly', 'monthly', 'yearly')),
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id)
		);

		CREATE TABLE IF NOT EXISTS recurring_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			currency_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT,
			frequency TEXT NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
			next_due_date DATE NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (category_id) REFERENCES categories (id),
			FOREIGN KEY (currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS user_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			primary_currency_id INTEGER NOT NULL,
			email_notifications BOOLEAN DEFAULT TRUE,
			budget_alerts BOOLEAN DEFAULT TRUE,
			recurring_reminders BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			FOREIGN KEY (primary_currency_id) REFERENCES currencies (id)
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			type TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		);
		`
	}

	_, err := db.Exec(createTablesSQL)
	return err
}

func createDefaultCategories(db *sql.DB) {
	// First, check if default categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE is_default = TRUE").Scan(&count)
	if err != nil {
		log.Printf("Error checking default categories: %v", err)
		return
	}

	// If default categories already exist, don't create them again
	if count > 0 {
		log.Println("Default categories already exist, skipping creation")
		return
	}

	// Default expense categories
	defaultExpenseCategories := []struct {
		name  string
		color string
	}{
		{"Food", "#EF4444"},
		{"Transport", "#3B82F6"},
		{"Entertainment", "#8B5CF6"},
		{"Shopping", "#F59E0B"},
		{"Bills", "#10B981"},
		{"Healthcare", "#EC4899"},
		{"Education", "#06B6D4"},
		{"Other", "#6B7280"},
	}

	// Default income categories
	defaultIncomeCategories := []struct {
		name  string
		color string
	}{
		{"Salary", "#10B981"},
		{"Bonus", "#F59E0B"},
		{"Freelance", "#8B5CF6"},
		{"Other", "#6B7280"},
	}

	// Create expense categories
	for _, cat := range defaultExpenseCategories {
		_, err := db.Exec(`
			INSERT INTO categories (name, color, is_default, category_type) 
			VALUES (?, ?, TRUE, 'expense')
		`, cat.name, cat.color)
		if err != nil {
			log.Printf("Error creating default expense category %s: %v", cat.name, err)
		}
	}

	// Create income categories
	for _, cat := range defaultIncomeCategories {
		_, err := db.Exec(`
			INSERT INTO categories (name, color, is_default, category_type) 
			VALUES (?, ?, TRUE, 'income')
		`, cat.name, cat.color)
		if err != nil {
			log.Printf("Error creating default income category %s: %v", cat.name, err)
		}
	}

	log.Printf("Created %d default expense categories and %d default income categories",
		len(defaultExpenseCategories), len(defaultIncomeCategories))
}

func createDefaultCurrencies(db *sql.DB) {
	// First, check if default currencies already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM currencies WHERE is_default = TRUE").Scan(&count)
	if err != nil {
		log.Printf("Error checking default currencies: %v", err)
		return
	}

	// If default currencies already exist, don't create them again
	if count > 0 {
		log.Println("Default currencies already exist, skipping creation")
		return
	}

	// Default currencies - 20 popular currencies
	defaultCurrencies := []struct {
		code   string
		name   string
		symbol string
	}{
		{"USD", "US Dollar", "$"},
		{"EUR", "Euro", "€"},
		{"GBP", "British Pound", "£"},
		{"JPY", "Japanese Yen", "¥"},
		{"CAD", "Canadian Dollar", "C$"},
		{"AUD", "Australian Dollar", "A$"},
		{"CHF", "Swiss Franc", "CHF"},
		{"CNY", "Chinese Yuan", "¥"},
		{"INR", "Indian Rupee", "₹"},
		{"BRL", "Brazilian Real", "R$"},
		{"KRW", "South Korean Won", "₩"},
		{"MXN", "Mexican Peso", "$"},
		{"SGD", "Singapore Dollar", "S$"},
		{"HKD", "Hong Kong Dollar", "HK$"},
		{"NZD", "New Zealand Dollar", "NZ$"},
		{"SEK", "Swedish Krona", "kr"},
		{"NOK", "Norwegian Krone", "kr"},
		{"DKK", "Danish Krone", "kr"},
		{"PLN", "Polish Złoty", "zł"},
		{"THB", "Thai Baht", "฿"},
	}

	// Create currencies
	for _, curr := range defaultCurrencies {
		_, err := db.Exec(`
			INSERT INTO currencies (code, name, symbol, is_default) 
			VALUES (?, ?, ?, TRUE)
		`, curr.code, curr.name, curr.symbol)
		if err != nil {
			log.Printf("Error creating default currency %s: %v", curr.code, err)
		}
	}

	log.Printf("Created %d default currencies", len(defaultCurrencies))
}
