package main

import (
	"log"
	"os"
	"panda-pocket/internal/application"
	"panda-pocket/internal/infrastructure/database"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize database with GORM
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Get underlying sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Create application with all dependencies
	app := application.NewApp(db)

	// Setup routes
	router := app.SetupRoutes()

	// Load .env (optional; continue if missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
