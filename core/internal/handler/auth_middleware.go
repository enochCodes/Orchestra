package handler

import (
	"strings"

	"github.com/enochcodes/orchestra/core/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddleware returns a middleware that validates JWT and sets user in context.
// When skipAuth is true (dev only), uses first admin user from DB.
func AuthMiddleware(db *gorm.DB, jwtSecret string, skipAuth bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if skipAuth {
			var user model.User
			if err := db.Where("system_role = ?", model.SystemRoleAdmin).First(&user).Error; err == nil {
				c.Locals("user", &user)
				c.Locals("user_id", user.ID)
			}
			// Even without user, allow request (some routes may not need user)
			return c.Next()
		}

		auth := c.Get("Authorization")
		if auth == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		var user model.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "user not found")
		}

		c.Locals("user", &user)
		c.Locals("user_id", user.ID)
		return c.Next()
	}
}

// RequireSystemAdmin ensures the user is a system admin.
func RequireSystemAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		user := u.(*model.User)
		if !user.IsSystemAdmin() {
			return fiber.NewError(fiber.StatusForbidden, "system admin access required")
		}
		return c.Next()
	}
}
