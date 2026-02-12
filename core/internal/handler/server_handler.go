package handler

import (
	"fmt"
	"strconv"

	tasks "github.com/enochcodes/orchestra/core/internal/engine"
	"github.com/enochcodes/orchestra/core/internal/model"
	"github.com/enochcodes/orchestra/core/internal/service"
	sshpkg "github.com/enochcodes/orchestra/core/pkg/ssh"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// ServerHandler handles HTTP requests for server management.
type ServerHandler struct {
	DB            *gorm.DB
	AsynqClient   *asynq.Client
	EncryptionKey string
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(db *gorm.DB, client *asynq.Client, encryptionKey string) *ServerHandler {
	return &ServerHandler{
		DB:            db,
		AsynqClient:   client,
		EncryptionKey: encryptionKey,
	}
}

// RegisterRequest represents the request body for server registration.
type RegisterRequest struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip" validate:"required"`
	SSHPort  int    `json:"ssh_port"`
	SSHUser  string `json:"ssh_user" validate:"required"`
	SSHKey   string `json:"ssh_key" validate:"required"`
}

// Register handles POST /api/v1/servers/register
func (h *ServerHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.IP == "" || req.SSHUser == "" || req.SSHKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "ip, ssh_user, and ssh_key are required")
	}

	// Default SSH port
	if req.SSHPort == 0 {
		req.SSHPort = 22
	}

	// Normalize PEM key (fixes paste issues: extra line breaks, wrong wraps)
	normalizedKey := sshpkg.NormalizePEMKey([]byte(req.SSHKey))

	// Encrypt SSH key
	encryptedKey, err := tasks.Encrypt(normalizedKey, h.EncryptionKey)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to encrypt SSH key")
	}

	// Get current user ID if authenticated
	var userID *uint
	if u := c.Locals("user"); u != nil {
		usr := u.(*model.User)
		userID = &usr.ID
	}

	// Create server record
	server := model.Server{
		Hostname:        req.Hostname,
		IP:              req.IP,
		SSHPort:         req.SSHPort,
		SSHUser:         req.SSHUser,
		SSHKeyEncrypted: encryptedKey,
		Status:          model.ServerStatusPending,
		CreatedByUserID: userID,
	}

	if err := h.DB.Create(&server).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to register server: %v", err))
	}

	// Enqueue pre-flight check task
	task, err := tasks.NewPreflightCheckTask(server.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create preflight task")
	}

	info, err := h.AsynqClient.Enqueue(task)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to enqueue preflight task")
	}

	_ = service.LogActivity(h.DB, model.ActivityTypeServerRegistered,
		fmt.Sprintf("Server %s (%s) registered", server.Hostname, server.IP),
		"server", server.ID, userID, nil)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":   "server registered, pre-flight check queued",
		"server_id": server.ID,
		"task_id":   info.ID,
	})
}

// List handles GET /api/v1/servers
func (h *ServerHandler) List(c *fiber.Ctx) error {
	var servers []model.Server
	if err := h.DB.Find(&servers).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch servers")
	}
	return c.JSON(fiber.Map{
		"servers": servers,
		"count":   len(servers),
	})
}

// Get handles GET /api/v1/servers/:id
func (h *ServerHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid server ID")
	}

	var server model.Server
	if err := h.DB.First(&server, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "server not found")
	}
	return c.JSON(server)
}

// GetLogs handles GET /api/v1/servers/:id/logs
func (h *ServerHandler) GetLogs(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid server ID")
	}

	var server model.Server
	if err := h.DB.First(&server, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "server not found")
	}

	return c.JSON(fiber.Map{
		"server_id":        server.ID,
		"status":           server.Status,
		"preflight_report": server.PreflightReport,
		"error_message":    server.ErrorMessage,
	})
}

// ListIdle handles GET /api/v1/servers/idle - servers not assigned to any cluster (role=none).
func (h *ServerHandler) ListIdle(c *fiber.Ctx) error {
	var servers []model.Server
	if err := h.DB.Where("role = ?", model.ServerRoleNone).
		Where("status = ?", model.ServerStatusReady).
		Find(&servers).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch idle servers")
	}
	return c.JSON(fiber.Map{
		"servers": servers,
		"count":   len(servers),
	})
}

// Update handles PATCH /api/v1/servers/:id
func (h *ServerHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid server ID")
	}

	var server model.Server
	if err := h.DB.First(&server, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "server not found")
	}

	var req struct {
		Hostname *string `json:"hostname"`
		TeamID   *uint   `json:"team_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Hostname != nil {
		server.Hostname = *req.Hostname
	}
	if req.TeamID != nil {
		server.TeamID = req.TeamID
	}

	if err := h.DB.Save(&server).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update server")
	}
	return c.JSON(server)
}

// ListTeams handles GET /api/v1/servers/teams
func (h *ServerHandler) ListTeams(c *fiber.Ctx) error {
	var teams []model.ServerTeam
	if err := h.DB.Find(&teams).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch teams")
	}
	return c.JSON(fiber.Map{
		"teams": teams,
		"count": len(teams),
	})
}

// CreateTeamRequest represents create team request.
type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTeam handles POST /api/v1/servers/teams
func (h *ServerHandler) CreateTeam(c *fiber.Ctx) error {
	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	team := model.ServerTeam{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := h.DB.Create(&team).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create team")
	}
	return c.Status(fiber.StatusCreated).JSON(team)
}

// Delete handles DELETE /api/v1/servers/:id
func (h *ServerHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid server ID")
	}
	if err := h.DB.Delete(&model.Server{}, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete server")
	}
	return c.JSON(fiber.Map{"message": "server deleted"})
}
