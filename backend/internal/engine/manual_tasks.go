package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/backend/internal/models"
	sshpkg "github.com/enochcodes/orchestra/backend/pkg/ssh"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const TypeManualClusterSetup = "cluster:manual_setup"

type ManualClusterPayload struct {
	ClusterID       uint   `json:"cluster_id"`
	ManagerServerID uint   `json:"manager_server_id"`
	WorkerServerIDs []uint `json:"worker_server_ids"`
}

func NewManualClusterSetupTask(clusterID, managerID uint, workerIDs []uint) (*asynq.Task, error) {
	payload, err := json.Marshal(ManualClusterPayload{
		ClusterID:       clusterID,
		ManagerServerID: managerID,
		WorkerServerIDs: workerIDs,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeManualClusterSetup, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

// ManualTaskHandler handles manual (no-orchestration) cluster setup.
type ManualTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

// HandleManualClusterSetup installs Docker on all nodes and marks the cluster active.
func (h *ManualTaskHandler) HandleManualClusterSetup(ctx context.Context, t *asynq.Task) error {
	var payload ManualClusterPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	log.Printf("Setting up manual cluster %d", payload.ClusterID)

	var cluster models.Cluster
	if err := h.DB.First(&cluster, payload.ClusterID).Error; err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	h.DB.Model(&cluster).Update("status", models.ClusterStatusProvisioning)

	// Install Docker on manager
	allServerIDs := append([]uint{payload.ManagerServerID}, payload.WorkerServerIDs...)
	for _, serverID := range allServerIDs {
		var server models.Server
		if err := h.DB.First(&server, serverID).Error; err != nil {
			log.Printf("Server %d not found: %v", serverID, err)
			continue
		}

		sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
		if err != nil {
			log.Printf("Failed to decrypt SSH key for server %d: %v", serverID, err)
			continue
		}

		client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
		if err != nil {
			log.Printf("SSH connection failed for server %d: %v", serverID, err)
			continue
		}

		// Install Docker
		installCmd := `command -v docker >/dev/null 2>&1 || { curl -fsSL https://get.docker.com | sh; }`
		result, err := client.ExecuteCommand(installCmd)
		if err != nil {
			log.Printf("Docker install failed on server %d: %s", serverID, result.Stderr)
			client.Close()
			continue
		}

		// Update role
		role := models.ServerRoleWorker
		if serverID == payload.ManagerServerID {
			role = models.ServerRoleManager
		}
		h.DB.Model(&server).Updates(map[string]interface{}{
			"role":       role,
			"cluster_id": cluster.ID,
		})

		client.Close()
		log.Printf("Docker installed on server %d for manual cluster %d", serverID, payload.ClusterID)
	}

	h.DB.Model(&cluster).Update("status", models.ClusterStatusActive)
	log.Printf("Manual cluster %d setup complete", payload.ClusterID)
	return nil
}
