package api

import (
	"fmt"
	"time"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/enochcodes/orchestra/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	DB         *gorm.DB
	JWTSecret  string
	JWTExpiry  time.Duration
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiry time.Duration) *AuthHandler {
	if jwtExpiry == 0 {
		jwtExpiry = 24 * time.Hour
	}
	return &AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
		JWTExpiry: jwtExpiry,
	}
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response.
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// UserResponse represents user data in API responses.
type UserResponse struct {
	ID          uint              `json:"id"`
	Email       string            `json:"email"`
	DisplayName string            `json:"display_name"`
	Avatar      string            `json:"avatar,omitempty"`
	SystemRole  models.SystemRole `json:"system_role"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID     uint   `json:"user_id"`
	Email      string `json:"email"`
	SystemRole string `json:"system_role"`
	jwt.RegisteredClaims
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
	}

	if !user.CheckPassword(req.Password) {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
	}

	token, expiresAt, err := h.generateToken(&user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate token")
	}

	// Log activity
	_ = services.LogActivity(h.DB, models.ActivityTypeUserLogin, fmt.Sprintf("User %s logged in", user.Email), "user", user.ID, &user.ID, nil)

	return c.JSON(LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			SystemRole:  user.SystemRole,
		},
	})
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SystemRole:  user.SystemRole,
	})
}

// UpdateProfile handles PATCH /api/v1/auth/me
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req struct {
		DisplayName *string `json:"display_name"`
		Avatar      *string `json:"avatar"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if err := h.DB.Save(user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update profile")
	}
	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SystemRole:  user.SystemRole,
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(h.JWTExpiry)
	claims := Claims{
		UserID:     user.ID,
		Email:      user.Email,
		SystemRole: string(user.SystemRole),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiresAt, nil
}
