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

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	currencies := []Currency{
		{"IDR", "Indonesian Rupiah", "Rp"},
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

	insertQuery := `
		INSERT INTO currencies (code, name, symbol, is_default, created_at, updated_at)
		SELECT $1, $2, $3, true, $4, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM currencies WHERE code = $1 AND user_id IS NULL
		)
	`

	now := time.Now()
	created := 0
	skipped := 0

	for _, currency := range currencies {
		result, err := db.Exec(insertQuery,
			currency.Code,
			currency.Name,
			currency.Symbol,
			now,
			now,
		)
		if err != nil {
			log.Printf("Failed to ensure currency %s: %v", currency.Code, err)
			continue
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			fmt.Printf("Added currency: %s (%s) - %s\n", currency.Code, currency.Name, currency.Symbol)
			created++
		} else {
			skipped++
		}
	}

	fmt.Printf("\nCreated %d missing currencies, %d already present\n", created, skipped)

	var finalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM currencies WHERE is_default = true AND user_id IS NULL").Scan(&finalCount)
	if err != nil {
		log.Printf("Warning: Failed to verify currency count: %v", err)
	} else {
		fmt.Printf("Total system default currencies in database: %d\n", finalCount)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
