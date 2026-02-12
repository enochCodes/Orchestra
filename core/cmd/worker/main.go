package main

import (
	"log"

	"github.com/enochcodes/orchestra/core/internal/config"
	"github.com/enochcodes/orchestra/core/internal/store"
	tasks "github.com/enochcodes/orchestra/core/internal/engine"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := store.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Worker: Database connected")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"provisioning": 6,
				"deployment":   3,
				"default":      1,
			},
		},
	)

	// SSH provisioning
	sshHandler := &tasks.SSHProvisionHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// K8s cluster tasks
	k8sHandler := &tasks.K8sTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// Docker Swarm tasks
	swarmHandler := &tasks.SwarmTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// Manual cluster tasks
	manualHandler := &tasks.ManualTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// Application deployment
	appHandler := &tasks.AppTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// Nginx provisioning
	nginxHandler := &tasks.NginxTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	// Environment push
	envHandler := &tasks.EnvTaskHandler{
		DB:            db,
		EncryptionKey: cfg.EncryptionKey,
	}

	mux := asynq.NewServeMux()

	// SSH
	mux.HandleFunc(tasks.TypePreflightCheck, sshHandler.HandlePreflightCheck)
	mux.HandleFunc(tasks.TypeInstallK3s, sshHandler.HandleInstallK3s)

	// K8s
	mux.HandleFunc(tasks.TypeDesignateManager, k8sHandler.HandleDesignateManager)
	mux.HandleFunc(tasks.TypeJoinWorker, k8sHandler.HandleJoinWorker)

	// Docker Swarm
	mux.HandleFunc(tasks.TypeSwarmInit, swarmHandler.HandleSwarmInit)
	mux.HandleFunc(tasks.TypeSwarmJoin, swarmHandler.HandleSwarmJoin)

	// Manual
	mux.HandleFunc(tasks.TypeManualClusterSetup, manualHandler.HandleManualClusterSetup)

	// Application
	mux.HandleFunc(tasks.TypeDeployApplication, appHandler.HandleDeployAppTask)

	// Nginx
	mux.HandleFunc(tasks.TypeNginxProvision, nginxHandler.HandleNginxProvision)

	// Environment
	mux.HandleFunc(tasks.TypePushEnv, envHandler.HandlePushEnv)

	log.Println("Orchestra Worker starting...")
	log.Println("  Tasks: preflight, k3s, swarm, manual, deploy, nginx, env")
	log.Println("  Queues: provisioning (6), deployment (3), default (1)")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
