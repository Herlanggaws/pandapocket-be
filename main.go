package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"panda-pocket/internal/application"
	appIdentity "panda-pocket/internal/application/identity"
	"panda-pocket/internal/infrastructure/analytics"
	"panda-pocket/internal/infrastructure/database"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	// Initialize database with GORM. InitDB loads .env before any connection.
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	analytics.InitFromEnv()
	defer analytics.Close()

	if err := initSentry(); err != nil {
		log.Printf("sentry init skipped: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	if err := appIdentity.RefuseDefaultTokenSecrets(); err != nil {
		log.Fatal(err)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func initSentry() error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return nil
	}
	sampleRate := 0.1
	if os.Getenv("GIN_MODE") != "release" {
		sampleRate = 1.0
	}
	environment := os.Getenv("SENTRY_ENVIRONMENT")
	if environment == "" {
		if os.Getenv("GIN_MODE") == "release" {
			environment = "production"
		} else {
			environment = "development"
		}
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          os.Getenv("SENTRY_RELEASE"),
		EnableTracing:    true,
		TracesSampleRate: sampleRate,
		SendDefaultPII:   false,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if event.Request != nil {
				event.Request.URL = scrubSensitiveURL(event.Request.URL)
				event.Request.QueryString = ""
				event.Request.Data = ""
				delete(event.Request.Headers, "Authorization")
				delete(event.Request.Headers, "Cookie")
			}
			return event
		},
	})
}

func scrubSensitiveURL(raw string) string {
	if raw == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		if i := strings.Index(raw, "?"); i >= 0 {
			return raw[:i]
		}
		return raw
	}
	q := parsed.Query()
	for _, key := range []string{"token", "description", "q", "password"} {
		q.Del(key)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}
