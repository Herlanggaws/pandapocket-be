package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Currency struct {
	Code   string
	Name   string
	Symbol string
}

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "herlangga.wicaksono")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "panda_pocket")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Define currencies including IDR
	currencies := []Currency{
		{"USD", "US Dollar", "$"},
		{"EUR", "Euro", "€"},
		{"GBP", "British Pound", "£"},
		{"JPY", "Japanese Yen", "¥"},
		{"AUD", "Australian Dollar", "A$"},
		{"CAD", "Canadian Dollar", "C$"},
		{"CHF", "Swiss Franc", "CHF"},
		{"CNY", "Chinese Yuan", "¥"},
		{"SEK", "Swedish Krona", "kr"},
		{"NOK", "Norwegian Krone", "kr"},
		{"DKK", "Danish Krone", "kr"},
		{"PLN", "Polish Zloty", "zł"},
		{"CZK", "Czech Koruna", "Kč"},
		{"HUF", "Hungarian Forint", "Ft"},
		{"RUB", "Russian Ruble", "₽"},
		{"BRL", "Brazilian Real", "R$"},
		{"INR", "Indian Rupee", "₹"},
		{"KRW", "South Korean Won", "₩"},
		{"SGD", "Singapore Dollar", "S$"},
		{"IDR", "Indonesian Rupiah", "Rp"},
	}

	// Check if currencies already exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM currencies WHERE is_default = true").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check existing currencies: %v", err)
	}

	if count > 0 {
		fmt.Printf("⚠️  Found %d existing default currencies. Skipping creation.\n", count)
		fmt.Println("To recreate currencies, delete existing ones first:")
		fmt.Println("DELETE FROM currencies WHERE is_default = true;")
		return
	}

	// Insert currencies
	insertQuery := `
		INSERT INTO currencies (code, name, symbol, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	successCount := 0

	for _, currency := range currencies {
		_, err := db.Exec(insertQuery,
			currency.Code,
			currency.Name,
			currency.Symbol,
			true, // is_default
			now,
			now,
		)

		if err != nil {
			log.Printf("❌ Failed to insert currency %s: %v", currency.Code, err)
		} else {
			fmt.Printf("✅ Added currency: %s (%s) - %s\n", currency.Code, currency.Name, currency.Symbol)
			successCount++
		}
	}

	fmt.Printf("\n🎉 Successfully added %d out of %d currencies\n", successCount, len(currencies))

	// Verify the insertion
	var finalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM currencies WHERE is_default = true").Scan(&finalCount)
	if err != nil {
		log.Printf("Warning: Failed to verify currency count: %v", err)
	} else {
		fmt.Printf("📊 Total default currencies in database: %d\n", finalCount)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
