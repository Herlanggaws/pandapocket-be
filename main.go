package main

import (
	"log"
	"panda-pocket/internal/application"
	"panda-pocket/internal/infrastructure/database"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create default categories and currencies
	database.CreateDefaultCategories(db)
	database.CreateDefaultCurrencies(db)

	// Create application with all dependencies
	app := application.NewApp(db)

	// Setup routes
	router := app.SetupRoutes()

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
