package handler

import (
	"fmt"
	"strconv"

	tasks "github.com/enochcodes/orchestra/core/internal/engine"
	"github.com/enochcodes/orchestra/core/internal/model"
	"github.com/enochcodes/orchestra/core/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type EnvironmentHandler struct {
	DB          *gorm.DB
	AsynqClient *asynq.Client
}

func NewEnvironmentHandler(db *gorm.DB, client *asynq.Client) *EnvironmentHandler {
	return &EnvironmentHandler{DB: db, AsynqClient: client}
}

// List returns all environments, optionally filtered by cluster_id
func (h *EnvironmentHandler) List(c *fiber.Ctx) error {
	var envs []model.Environment
	query := h.DB.Preload("Cluster")
	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if err := query.Find(&envs).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch environments")
	}
	return c.JSON(fiber.Map{"environments": envs, "count": len(envs)})
}

// Get returns a single environment by ID
func (h *EnvironmentHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid environment ID")
	}
	var env model.Environment
	if err := h.DB.Preload("Cluster").First(&env, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "environment not found")
	}
	return c.JSON(env)
}

// Create creates a new environment config
func (h *EnvironmentHandler) Create(c *fiber.Ctx) error {
	var req struct {
		ClusterID uint             `json:"cluster_id"`
		Scope     string           `json:"scope"`
		Name      string           `json:"name"`
		Variables model.EnvVarMap `json:"variables"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}
	if req.ClusterID == 0 || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "cluster_id and name are required")
	}
	scope := model.EnvScope(req.Scope)
	if scope == "" {
		scope = model.EnvScopeProduction
	}

	env := model.Environment{
		ClusterID: req.ClusterID,
		Scope:     scope,
		Name:      req.Name,
		Variables: req.Variables,
	}
	if err := h.DB.Create(&env).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create environment")
	}

	var userID *uint
	if u := c.Locals("user"); u != nil {
		usr := u.(*model.User)
		userID = &usr.ID
	}
	_ = service.LogActivity(h.DB, model.ActivityTypeEnvPushed,
		fmt.Sprintf("Environment '%s' created for cluster %d", env.Name, env.ClusterID),
		"environment", env.ID, userID, nil)

	return c.Status(fiber.StatusCreated).JSON(env)
}

// Update updates environment variables
func (h *EnvironmentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}
	var env model.Environment
	if err := h.DB.First(&env, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}

	var req struct {
		Name      *string           `json:"name"`
		Variables *model.EnvVarMap `json:"variables"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}
	if req.Name != nil {
		env.Name = *req.Name
	}
	if req.Variables != nil {
		env.Variables = *req.Variables
		env.Synced = false
	}
	if err := h.DB.Save(&env).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update")
	}
	return c.JSON(env)
}

// Push pushes environment variables to all servers in the cluster
func (h *EnvironmentHandler) Push(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}

	task, err := tasks.NewPushEnvTask(uint(id))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create push task")
	}
	if _, err := h.AsynqClient.Enqueue(task); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to enqueue push task")
	}

	return c.JSON(fiber.Map{"message": "environment push queued"})
}

// Delete removes an environment
func (h *EnvironmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}
	if err := h.DB.Delete(&model.Environment{}, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete")
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
