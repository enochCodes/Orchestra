package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/backend/internal/buildpack"
	"github.com/enochcodes/orchestra/backend/internal/models"
	"gorm.io/gorm"

	"github.com/hibiken/asynq"
)

const (
	TypeDeployApplication = "app:deploy"
)

type DeployAppPayload struct {
	AppID uint `json:"app_id"`
}

type AppTaskHandler struct {
	DB *gorm.DB
}

func NewDeployAppTask(appID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(DeployAppPayload{AppID: appID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDeployApplication, payload), nil
}

func (h *AppTaskHandler) HandleDeployAppTask(ctx context.Context, t *asynq.Task) error {
	var p DeployAppPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Starting deployment for App ID: %d", p.AppID)

	var app models.Application
	if err := h.DB.Preload("Cluster").First(&app, p.AppID).Error; err != nil {
		return fmt.Errorf("app lookup failed: %v", err)
	}

	// 1. Create a Deployment record
	deployment := models.Deployment{
		ApplicationID: app.ID,
		Version:       "v1.0.0", // Simple versioning for now
		Status:        models.DeploymentStatusBuilding,
	}
	if err := h.DB.Create(&deployment).Error; err != nil {
		return fmt.Errorf("failed to create deployment record: %v", err)
	}

	// 2. Generate Dockerfile
	dockerfile := buildpack.GenerateDockerfile(app.BuildType, app.BuildCmd, app.StartCmd)
	log.Printf("Generated Dockerfile for %s (%s):\n%s", app.Name, app.BuildType, dockerfile)

	// 3. Update status to Deploying
	h.DB.Model(&deployment).Update("status", models.DeploymentStatusDeploying)
	h.DB.Model(&app).Update("status", "deploying")

	// 4. Simulate Build Process
	// In reality: SSH to cluster manager -> Write Dockerfile -> Docker Build -> Docker Push
	log.Printf("Simulating build on cluster %s...", app.Cluster.Name)

	// 5. Update Status to Live
	h.DB.Model(&deployment).Update("status", models.DeploymentStatusLive)
	h.DB.Model(&app).Update("status", "running")
	log.Printf("Deployment successful for App %s", app.Name)

	return nil
}
