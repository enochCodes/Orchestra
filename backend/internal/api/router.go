package api

import (
	"time"

	"github.com/enochcodes/orchestra/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// SetupRouter configures the Fiber app with middleware and routes.
func SetupRouter(db *gorm.DB, asynqClient *asynq.Client, encryptionKey, jwtSecret string, skipAuth bool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Orchestra API",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))

	// Health check (public)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "orchestra-api",
		})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Auth routes (login is public)
	authHandler := NewAuthHandler(db, jwtSecret, 24*time.Hour)
	v1.Post("/auth/login", authHandler.Login)

	// Protected routes - require JWT (or skip in dev)
	auth := v1.Group("", AuthMiddleware(db, jwtSecret, skipAuth))
	auth.Get("/auth/me", authHandler.Me)
	auth.Patch("/auth/me", authHandler.UpdateProfile)

	// Admin-only routes (system admin)
	admin := auth.Group("", RequireSystemAdmin())
	userHandler := NewUserHandler(db)
	admin.Get("/users", userHandler.List)
	admin.Post("/users", userHandler.Create)
	admin.Get("/users/:id", userHandler.Get)
	admin.Patch("/users/:id", userHandler.Update)

	// Server routes
	serverHandler := NewServerHandler(db, asynqClient, encryptionKey)
	servers := auth.Group("/servers")
	servers.Post("/register", serverHandler.Register)
	servers.Get("/", serverHandler.List)
	servers.Get("/idle", serverHandler.ListIdle)
	servers.Get("/teams", serverHandler.ListTeams)
	servers.Post("/teams", serverHandler.CreateTeam)
	servers.Get("/:id", serverHandler.Get)
	servers.Get("/:id/logs", serverHandler.GetLogs)
	servers.Patch("/:id", serverHandler.Update)
	servers.Delete("/:id", serverHandler.Delete)

	// Cluster routes
	clusterSvc := services.NewClusterService(db, asynqClient, encryptionKey)
	clusterHandler := NewClusterHandler(clusterSvc)
	clusters := auth.Group("/clusters")
	clusters.Post("/design", clusterHandler.Design)
	clusters.Get("/", clusterHandler.List)
	clusters.Get("/:id", clusterHandler.Get)

	// Metadata routes
	metadata := auth.Group("/metadata")
	metadata.Get("/frameworks", GetFrameworks)
	metadata.Get("/stacks", GetStacks)

	// Application routes
	appHandler := NewApplicationHandler(db, asynqClient)
	applications := auth.Group("/applications")
	applications.Get("/", appHandler.List)
	applications.Post("/", appHandler.Create)
	applications.Get("/:id", appHandler.Get)
	applications.Patch("/:id", appHandler.Update)
	applications.Delete("/:id", appHandler.Delete)
	applications.Post("/:id/redeploy", appHandler.Redeploy)

	// Deployment routes
	depHandler := NewDeploymentHandler(db)
	deployments := auth.Group("/deployments")
	deployments.Get("/", depHandler.List)
	deployments.Get("/:id", depHandler.Get)
	deployments.Get("/:id/logs", depHandler.GetLogs)

	// Environment routes
	envHandler := NewEnvironmentHandler(db, asynqClient)
	environments := auth.Group("/environments")
	environments.Get("/", envHandler.List)
	environments.Post("/", envHandler.Create)
	environments.Get("/:id", envHandler.Get)
	environments.Patch("/:id", envHandler.Update)
	environments.Delete("/:id", envHandler.Delete)
	environments.Post("/:id/push", envHandler.Push)

	// Nginx routes
	nginxHandler := NewNginxHandler(db, asynqClient)
	nginx := auth.Group("/nginx")
	nginx.Get("/", nginxHandler.List)
	nginx.Post("/", nginxHandler.Create)
	nginx.Delete("/:id", nginxHandler.Delete)

	// Activity routes
	activityHandler := NewActivityHandler(db)
	auth.Get("/activities", activityHandler.List)

	// Monitoring routes
	monHandler := NewMonitoringHandler(db)
	monitoring := auth.Group("/monitoring")
	monitoring.Get("/overview", monHandler.GetOverview)
	monitoring.Get("/status", monHandler.GetSystemStatus)
	monitoring.Get("/infra", monHandler.GetInfraDetails)

	return app
}

// customErrorHandler provides consistent error responses.
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}
