package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"panda-pocket/internal/application"
	"panda-pocket/internal/infrastructure/database"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app.StartBackgroundJobs(ctx)

	// Setup routes
	router := app.SetupRoutes()

	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
