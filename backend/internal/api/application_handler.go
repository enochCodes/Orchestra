package api

import (
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/enochcodes/orchestra/backend/internal/services"
	"github.com/enochcodes/orchestra/backend/internal/tasks"
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
	if err := h.DB.Preload("Cluster").Find(&apps).Error; err != nil {
		return err
	}
	return c.JSON(apps)
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
		Name        *string               `json:"name"`
		Replicas    *int                  `json:"replicas"`
		BuildCmd    *string              `json:"build_cmd"`
		StartCmd    *string              `json:"start_cmd"`
		EnvVars     *models.ScopedEnvs   `json:"env_vars"`
		Status      *string              `json:"status"`
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

	// Default status
	app.Status = "pending"
	if app.Replicas == 0 {
		app.Replicas = 1
	}

	// Save to DB
	if err := h.DB.Create(&app).Error; err != nil {
		log.Printf("Failed to create app: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create application")
	}

	// Trigger Deployment Task
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
