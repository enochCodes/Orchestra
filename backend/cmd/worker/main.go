package main

import (
	"log"

	"github.com/enochcodes/orchestra/backend/internal/config"
	"github.com/enochcodes/orchestra/backend/internal/database"
	"github.com/enochcodes/orchestra/backend/internal/tasks"
	"github.com/hibiken/asynq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database (workers need DB access for task handlers)
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Worker: Database connected")

	// Initialize Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"provisioning": 6, // SSH tasks (high latency, prioritize)
				"deployment":   3, // K8s API tasks (low latency)
				"default":      1,
			},
		},
	)

	// Register task handlers
	sshHandler := &tasks.SSHProvisionHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	k8sHandler := &tasks.K8sTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// App tasks
	appHandler := &tasks.AppTaskHandler{
		DB: db,
	}

	mux := asynq.NewServeMux()

	// SSH provisioning tasks
	mux.HandleFunc(tasks.TypePreflightCheck, sshHandler.HandlePreflightCheck)
	mux.HandleFunc(tasks.TypeInstallK3s, sshHandler.HandleInstallK3s)

	// K8s cluster tasks
	mux.HandleFunc(tasks.TypeDesignateManager, k8sHandler.HandleDesignateManager)
	mux.HandleFunc(tasks.TypeJoinWorker, k8sHandler.HandleJoinWorker)

	// Application Deployment
	mux.HandleFunc(tasks.TypeDeployApplication, appHandler.HandleDeployAppTask)

	log.Println("Orchestra Worker starting...")
	log.Println("  Queues: provisioning (6), deployment (3), default (1)")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
