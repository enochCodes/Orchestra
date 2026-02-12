package handler

import (
	"strconv"

	tasks "github.com/enochcodes/orchestra/core/internal/engine"
	"github.com/enochcodes/orchestra/core/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type NginxHandler struct {
	DB          *gorm.DB
	AsynqClient *asynq.Client
}

func NewNginxHandler(db *gorm.DB, client *asynq.Client) *NginxHandler {
	return &NginxHandler{DB: db, AsynqClient: client}
}

// List returns all nginx configs, optionally by server_id
func (h *NginxHandler) List(c *fiber.Ctx) error {
	var configs []model.NginxConfig
	query := h.DB.Preload("Server")
	if serverID := c.Query("server_id"); serverID != "" {
		query = query.Where("server_id = ?", serverID)
	}
	if err := query.Find(&configs).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch nginx configs")
	}
	return c.JSON(fiber.Map{"configs": configs, "count": len(configs)})
}

// Create creates a new nginx config and provisions it
func (h *NginxHandler) Create(c *fiber.Ctx) error {
	var req struct {
		ServerID      uint   `json:"server_id"`
		Domain        string `json:"domain"`
		UpstreamPort  int    `json:"upstream_port"`
		SSLEnabled    bool   `json:"ssl_enabled"`
		LetsEncrypt   bool   `json:"lets_encrypt"`
		CustomConfig  string `json:"custom_config"`
		ApplicationID *uint  `json:"application_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request")
	}
	if req.ServerID == 0 || req.Domain == "" || req.UpstreamPort == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "server_id, domain, and upstream_port required")
	}

	cfg := model.NginxConfig{
		ServerID:      req.ServerID,
		Domain:        req.Domain,
		UpstreamPort:  req.UpstreamPort,
		SSLEnabled:    req.SSLEnabled,
		LetsEncrypt:   req.LetsEncrypt,
		CustomConfig:  req.CustomConfig,
		ApplicationID: req.ApplicationID,
		Status:        "pending",
	}
	if err := h.DB.Create(&cfg).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create config")
	}

	// Enqueue provisioning
	task, err := tasks.NewNginxProvisionTask(cfg.ID)
	if err == nil {
		h.AsynqClient.Enqueue(task)
	}

	return c.Status(fiber.StatusCreated).JSON(cfg)
}

// Delete removes a nginx config
func (h *NginxHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}
	if err := h.DB.Delete(&model.NginxConfig{}, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete")
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
