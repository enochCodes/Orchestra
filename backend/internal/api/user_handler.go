package api

import (
	"strconv"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UserHandler handles user management (system admin only).
type UserHandler struct {
	DB *gorm.DB
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// List handles GET /api/v1/users
func (h *UserHandler) List(c *fiber.Ctx) error {
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch users")
	}

	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Avatar:      u.Avatar,
			SystemRole:  u.SystemRole,
		}
	}
	return c.JSON(fiber.Map{
		"users": result,
		"count": len(result),
	})
}

// CreateRequest represents create user request.
type CreateUserRequest struct {
	Email       string           `json:"email"`
	Password    string           `json:"password"`
	DisplayName string           `json:"display_name"`
	SystemRole  models.SystemRole `json:"system_role"`
}

// Create handles POST /api/v1/users
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	user := models.User{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		SystemRole:  req.SystemRole,
	}
	if user.SystemRole == "" {
		user.SystemRole = models.SystemRoleDeveloper
	}
	if err := user.SetPassword(req.Password); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to set password")
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create user")
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SystemRole:  user.SystemRole,
	})
}

// Get handles GET /api/v1/users/:id
func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	var user models.User
	if err := h.DB.First(&user, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SystemRole:  user.SystemRole,
	})
}

// Update handles PATCH /api/v1/users/:id
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	var user models.User
	if err := h.DB.First(&user, uint(id)).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	var req struct {
		DisplayName *string           `json:"display_name"`
		SystemRole  *models.SystemRole `json:"system_role"`
		Password    *string           `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.SystemRole != nil {
		user.SystemRole = *req.SystemRole
	}
	if req.Password != nil && *req.Password != "" {
		if err := user.SetPassword(*req.Password); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to set password")
		}
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update user")
	}

	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SystemRole:  user.SystemRole,
	})
}

