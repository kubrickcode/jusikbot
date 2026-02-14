package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jusikbot/collector/internal/store"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := store.ConnectDB(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := store.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("migrations applied successfully")
}
