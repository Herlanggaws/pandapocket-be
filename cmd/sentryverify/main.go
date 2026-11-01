package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

// One-shot verification using the same ClientOptions as production main.
func main() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Fatal("SENTRY_DSN required")
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      "development",
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		SendDefaultPII:   false,
	}); err != nil {
		log.Fatal(err)
	}
	defer sentry.Flush(5 * time.Second)

	sentry.CaptureException(fmt.Errorf("Bear Budget API Sentry verification error"))
	log.Println("sent verification error")
}
