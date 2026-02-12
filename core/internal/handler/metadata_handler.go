package handler

import (
	"github.com/enochcodes/orchestra/core/internal/buildpack"
	"github.com/gofiber/fiber/v2"
)

func GetFrameworks(c *fiber.Ctx) error {
	return c.JSON(buildpack.GetMetadata())
}

// GetStacks returns available deployment stacks (same as frameworks for now).
func GetStacks(c *fiber.Ctx) error {
	return c.JSON(buildpack.GetMetadata())
}
