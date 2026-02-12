package api

import (
	"fmt"
	"log"
	"strconv"

	tasks "github.com/enochcodes/orchestra/backend/internal/engine"
	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/enochcodes/orchestra/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type ApplicationHandler struct {
	DB          *gorm.DB
	AsynqClient *asynq.Client
}

func NewApplicationHandler(db *gorm.DB, client *asynq.Client) *ApplicationHandler {
	return &ApplicationHandler{
		DB:          db,
		AsynqClient: client,
	}
}

// List applications
func (h *ApplicationHandler) List(c *fiber.Ctx) error {
	var apps []models.Application
	query := h.DB.Preload("Cluster")
	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if err := query.Find(&apps).Error; err != nil {
		return err
	}
	return c.JSON(fiber.Map{"applications": apps, "count": len(apps)})
}

// Get application by ID
func (h *ApplicationHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	var app models.Application
	if err := h.DB.Preload("Cluster").First(&app, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Application not found")
	}
	return c.JSON(app)
}

// Update application
func (h *ApplicationHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var app models.Application
	if err := h.DB.Preload("Cluster").First(&app, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Application not found")
	}

	var req struct {
		Name     *string            `json:"name"`
		Replicas *int               `json:"replicas"`
		BuildCmd *string            `json:"build_cmd"`
		StartCmd *string            `json:"start_cmd"`
		EnvVars  *models.ScopedEnvs `json:"env_vars"`
		Status   *string            `json:"status"`
		Port     *int               `json:"port"`
		Domain   *string            `json:"domain"`
		Branch   *string            `json:"branch"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name != nil {
		app.Name = *req.Name
	}
	if req.Replicas != nil {
		app.Replicas = *req.Replicas
	}
	if req.BuildCmd != nil {
		app.BuildCmd = *req.BuildCmd
	}
	if req.StartCmd != nil {
		app.StartCmd = *req.StartCmd
	}
	if req.EnvVars != nil {
		app.EnvVars = *req.EnvVars
	}
	if req.Status != nil {
		app.Status = *req.Status
	}
	if req.Port != nil {
		app.Port = *req.Port
	}
	if req.Domain != nil {
		app.Domain = *req.Domain
	}
	if req.Branch != nil {
		app.Branch = *req.Branch
	}

	if err := h.DB.Save(&app).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update application")
	}
	return c.JSON(app)
}

// Create application and trigger deployment
func (h *ApplicationHandler) Create(c *fiber.Ctx) error {
	var app models.Application
	if err := c.BodyParser(&app); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	app.Status = "pending"
	if app.Replicas == 0 {
		app.Replicas = 1
	}
	if app.Namespace == "" {
		app.Namespace = "default"
	}

	if err := h.DB.Create(&app).Error; err != nil {
		log.Printf("Failed to create app: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create application")
	}

	// Trigger deployment
	task, err := tasks.NewDeployAppTask(app.ID)
	if err != nil {
		log.Printf("Failed to create deploy task: %v", err)
	} else {
		if _, err := h.AsynqClient.Enqueue(task); err != nil {
			log.Printf("Failed to enqueue deploy task: %v", err)
		}
	}

	var userID *uint
	if u := c.Locals("user"); u != nil {
		usr := u.(*models.User)
		userID = &usr.ID
	}
	_ = services.LogActivity(h.DB, models.ActivityTypeAppDeployed,
		fmt.Sprintf("Application '%s' deployment initiated", app.Name),
		"application", app.ID, userID, nil)

	return c.Status(fiber.StatusCreated).JSON(app)
}

// Redeploy triggers a new deployment for an existing application
func (h *ApplicationHandler) Redeploy(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}

	var app models.Application
	if err := h.DB.First(&app, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Application not found")
	}

	task, err := tasks.NewDeployAppTask(app.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create deploy task")
	}
	if _, err := h.AsynqClient.Enqueue(task); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to enqueue deploy task")
	}

	h.DB.Model(&app).Update("status", "pending")

	var userID *uint
	if u := c.Locals("user"); u != nil {
		usr := u.(*models.User)
		userID = &usr.ID
	}
	_ = services.LogActivity(h.DB, models.ActivityTypeAppRedeployed,
		fmt.Sprintf("Application '%s' redeployment triggered", app.Name),
		"application", app.ID, userID, nil)

	return c.JSON(fiber.Map{"message": "redeployment queued"})
}

// Delete removes an application
func (h *ApplicationHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}
	if err := h.DB.Delete(&models.Application{}, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete")
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
