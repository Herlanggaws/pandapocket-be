package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Panda Pocket PostgreSQL Database Setup")
	fmt.Println("=====================================")

	// Connect to PostgreSQL database
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
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

	fmt.Println("Connected to PostgreSQL database successfully!")
	fmt.Println("Database setup completed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
