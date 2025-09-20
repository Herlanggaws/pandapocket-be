package database

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

// InitDB initializes the PostgreSQL database connection using GORM
func InitDB() (*gorm.DB, error) {
	// Load environment variables from .env file
	loadEnvFile()

	db, err := initGormDB()
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = autoMigrate(db)
	if err != nil {
		return nil, err
	}

	// Create default data
	err = createDefaultData(db)
	if err != nil {
		return nil, err
	}

	log.Printf("Database initialized successfully with GORM and PostgreSQL")
	return db, nil
}

func initGormDB() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "herlangga.wicaksono")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "panda_pocket")

	var dsn string
	if password != "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	} else {
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
			host, port, user, dbname)
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// autoMigrate runs GORM auto-migration for all models
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Currency{},
		&Category{},
		&Expense{},
		&Income{},
		&Budget{},
		&RecurringTransaction{},
		&UserPreferences{},
		&Notification{},
	)
}

// createDefaultData creates default categories and currencies using GORM
func createDefaultData(db *gorm.DB) error {
	// Create default categories
	err := createDefaultCategoriesGorm(db)
	if err != nil {
		return err
	}

	// Create default currencies
	err = createDefaultCurrenciesGorm(db)
	if err != nil {
		return err
	}

	return nil
}

// createDefaultCategoriesGorm creates default categories using GORM
func createDefaultCategoriesGorm(db *gorm.DB) error {
	// Check if default categories already exist
	var count int64
	err := db.Model(&Category{}).Where("is_default = ?", true).Count(&count).Error
	if err != nil {
		return err
	}

	// If default categories already exist, don't create them again
	if count > 0 {
		log.Println("Default categories already exist, skipping creation")
		return nil
	}

	// Default expense categories
	defaultExpenseCategories := []Category{
		{Name: "Food", Color: "#EF4444", IsDefault: true, CategoryType: "expense"},
		{Name: "Transport", Color: "#3B82F6", IsDefault: true, CategoryType: "expense"},
		{Name: "Entertainment", Color: "#8B5CF6", IsDefault: true, CategoryType: "expense"},
		{Name: "Shopping", Color: "#F59E0B", IsDefault: true, CategoryType: "expense"},
		{Name: "Bills", Color: "#10B981", IsDefault: true, CategoryType: "expense"},
		{Name: "Healthcare", Color: "#EC4899", IsDefault: true, CategoryType: "expense"},
		{Name: "Education", Color: "#06B6D4", IsDefault: true, CategoryType: "expense"},
		{Name: "Other", Color: "#6B7280", IsDefault: true, CategoryType: "expense"},
	}

	// Default income categories
	defaultIncomeCategories := []Category{
		{Name: "Salary", Color: "#10B981", IsDefault: true, CategoryType: "income"},
		{Name: "Bonus", Color: "#F59E0B", IsDefault: true, CategoryType: "income"},
		{Name: "Freelance", Color: "#8B5CF6", IsDefault: true, CategoryType: "income"},
		{Name: "Other", Color: "#6B7280", IsDefault: true, CategoryType: "income"},
	}

	// Create expense categories
	if err := db.Create(&defaultExpenseCategories).Error; err != nil {
		return err
	}

	// Create income categories
	if err := db.Create(&defaultIncomeCategories).Error; err != nil {
		return err
	}

	log.Printf("Created %d default expense categories and %d default income categories",
		len(defaultExpenseCategories), len(defaultIncomeCategories))
	return nil
}

// createDefaultCurrenciesGorm creates default currencies using GORM
func createDefaultCurrenciesGorm(db *gorm.DB) error {
	// Check if default currencies already exist
	var count int64
	err := db.Model(&Currency{}).Where("is_default = ?", true).Count(&count).Error
	if err != nil {
		return err
	}

	// If default currencies already exist, don't create them again
	if count > 0 {
		log.Println("Default currencies already exist, skipping creation")
		return nil
	}

	// Default currencies - 20 popular currencies
	defaultCurrencies := []Currency{
		{Code: "USD", Name: "US Dollar", Symbol: "$", IsDefault: true},
		{Code: "EUR", Name: "Euro", Symbol: "€", IsDefault: true},
		{Code: "GBP", Name: "British Pound", Symbol: "£", IsDefault: true},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", IsDefault: true},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", IsDefault: true},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", IsDefault: true},
		{Code: "CHF", Name: "Swiss Franc", Symbol: "CHF", IsDefault: true},
		{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", IsDefault: true},
		{Code: "INR", Name: "Indian Rupee", Symbol: "₹", IsDefault: true},
		{Code: "BRL", Name: "Brazilian Real", Symbol: "R$", IsDefault: true},
		{Code: "KRW", Name: "South Korean Won", Symbol: "₩", IsDefault: true},
		{Code: "MXN", Name: "Mexican Peso", Symbol: "$", IsDefault: true},
		{Code: "SGD", Name: "Singapore Dollar", Symbol: "S$", IsDefault: true},
		{Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$", IsDefault: true},
		{Code: "NZD", Name: "New Zealand Dollar", Symbol: "NZ$", IsDefault: true},
		{Code: "SEK", Name: "Swedish Krona", Symbol: "kr", IsDefault: true},
		{Code: "NOK", Name: "Norwegian Krone", Symbol: "kr", IsDefault: true},
		{Code: "DKK", Name: "Danish Krone", Symbol: "kr", IsDefault: true},
		{Code: "PLN", Name: "Polish Złoty", Symbol: "zł", IsDefault: true},
		{Code: "THB", Name: "Thai Baht", Symbol: "฿", IsDefault: true},
	}

	// Create currencies
	if err := db.Create(&defaultCurrencies).Error; err != nil {
		return err
	}

	log.Printf("Created %d default currencies", len(defaultCurrencies))
	return nil
}
