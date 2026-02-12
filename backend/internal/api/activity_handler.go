package api

import (
	"strconv"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ActivityHandler handles activity/audit log endpoints.
type ActivityHandler struct {
	DB *gorm.DB
}

// NewActivityHandler creates a new ActivityHandler.
func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{DB: db}
}

// List handles GET /api/v1/activities
func (h *ActivityHandler) List(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	var activities []models.Activity
	if err := h.DB.Order("created_at DESC").Limit(limit).Find(&activities).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch activities")
	}

	return c.JSON(fiber.Map{
		"activities": activities,
		"count":      len(activities),
	})
}
