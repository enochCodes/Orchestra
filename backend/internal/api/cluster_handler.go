package api

import (
	"fmt"
	"strconv"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"github.com/enochcodes/orchestra/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// ClusterHandler handles HTTP requests for cluster management.
type ClusterHandler struct {
	Service *services.ClusterService
}

// NewClusterHandler creates a new ClusterHandler.
func NewClusterHandler(svc *services.ClusterService) *ClusterHandler {
	return &ClusterHandler{Service: svc}
}

// Design handles POST /api/v1/clusters/design
func (h *ClusterHandler) Design(c *fiber.Ctx) error {
	var input services.DesignClusterInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.Name == "" || input.ManagerServerID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "name and manager_server_id are required")
	}

	cluster, err := h.Service.DesignCluster(input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var userID *uint
	if u := c.Locals("user"); u != nil {
		usr := u.(*models.User)
		userID = &usr.ID
	}
	_ = services.LogActivity(h.Service.DB, models.ActivityTypeClusterCreated,
		fmt.Sprintf("Cluster '%s' design initiated", cluster.Name),
		"cluster", cluster.ID, userID, nil)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":    "cluster design accepted, provisioning started",
		"cluster_id": cluster.ID,
	})
}

// List handles GET /api/v1/clusters
func (h *ClusterHandler) List(c *fiber.Ctx) error {
	clusters, err := h.Service.ListClusters()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"clusters": clusters,
		"count":    len(clusters),
	})
}

// Get handles GET /api/v1/clusters/:id
func (h *ClusterHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cluster ID")
	}

	cluster, err := h.Service.GetCluster(uint(id))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(cluster)
}
