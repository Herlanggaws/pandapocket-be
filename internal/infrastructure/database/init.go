package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, use system environment variables
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// InitDB initializes the PostgreSQL database connection
func InitDB() (*sql.DB, error) {
	// Load environment variables from .env file
	loadEnvFile()

	db, err := initPostgresDB()
	if err != nil {
		return nil, err
	}

	// Create tables
	err = createTables(db, "postgres")
	if err != nil {
		return nil, err
	}

	log.Printf("Database initialized successfully with postgres")
	return db, nil
}

func initPostgresDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "herlangga.wicaksono")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "panda_pocket")

	var psqlInfo string
	if password != "" {
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
			host, port, user, dbname)
	}

	return sql.Open("postgres", psqlInfo)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTables(db *sql.DB, dbType string) error {
	createTablesSQL := `
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

	_, err := db.Exec(createTablesSQL)
	return err
}

// CreateDefaultCategories creates default categories
func CreateDefaultCategories(db *sql.DB) {
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
			VALUES ($1, $2, TRUE, 'expense')
		`, cat.name, cat.color)
		if err != nil {
			log.Printf("Error creating default expense category %s: %v", cat.name, err)
		}
	}

	// Create income categories
	for _, cat := range defaultIncomeCategories {
		_, err := db.Exec(`
			INSERT INTO categories (name, color, is_default, category_type) 
			VALUES ($1, $2, TRUE, 'income')
		`, cat.name, cat.color)
		if err != nil {
			log.Printf("Error creating default income category %s: %v", cat.name, err)
		}
	}

	log.Printf("Created %d default expense categories and %d default income categories",
		len(defaultExpenseCategories), len(defaultIncomeCategories))
}

// CreateDefaultCurrencies creates default currencies
func CreateDefaultCurrencies(db *sql.DB) {
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
			VALUES ($1, $2, $3, TRUE)
		`, curr.code, curr.name, curr.symbol)
		if err != nil {
			log.Printf("Error creating default currency %s: %v", curr.code, err)
		}
	}

	log.Printf("Created %d default currencies", len(defaultCurrencies))
}
