package main

import (
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/backend/internal/api"
	"github.com/enochcodes/orchestra/backend/internal/config"
	"github.com/enochcodes/orchestra/backend/internal/database"
	"github.com/hibiken/asynq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected and migrated")

	// Seed initial data (system admin)
	if err := database.Seed(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer asynqClient.Close()
	log.Println("Asynq client initialized")

	// Setup Fiber router
	app := api.SetupRouter(db, asynqClient, cfg.EncryptionKey, cfg.JWTSecret, cfg.SkipAuth)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Orchestra API server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
