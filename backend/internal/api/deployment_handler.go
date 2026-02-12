package api

import (
	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DeploymentHandler struct {
	DB *gorm.DB
}

func NewDeploymentHandler(db *gorm.DB) *DeploymentHandler {
	return &DeploymentHandler{DB: db}
}

// List deployments
func (h *DeploymentHandler) List(c *fiber.Ctx) error {
	var deployments []models.Deployment
	if err := h.DB.Preload("Application").Order("created_at desc").Find(&deployments).Error; err != nil {
		return err
	}
	return c.JSON(deployments)
}

// Get deployment by ID
func (h *DeploymentHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	var deployment models.Deployment
	if err := h.DB.Preload("Application").First(&deployment, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Deployment not found")
	}
	return c.JSON(deployment)
}

// GetLogs handles GET /api/v1/deployments/:id/logs
func (h *DeploymentHandler) GetLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	var deployment models.Deployment
	if err := h.DB.Preload("Application").First(&deployment, id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Deployment not found")
	}
	return c.JSON(fiber.Map{
		"deployment_id": deployment.ID,
		"application":   deployment.Application.Name,
		"version":       deployment.Version,
		"status":        deployment.Status,
		"logs":          deployment.Logs,
		"created_at":    deployment.CreatedAt,
	})
}
