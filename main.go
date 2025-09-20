package main

import (
	"log"
	"panda-pocket/internal/application"
	"panda-pocket/internal/infrastructure/database"
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

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
