package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/enochcodes/orchestra/core/internal/model"
	sshpkg "github.com/enochcodes/orchestra/core/pkg/ssh"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const (
	TypeSwarmInit = "cluster:swarm_init"
	TypeSwarmJoin = "cluster:swarm_join"
)

type SwarmInitPayload struct {
	ClusterID uint `json:"cluster_id"`
	ServerID  uint `json:"server_id"`
}

type SwarmJoinPayload struct {
	ClusterID uint `json:"cluster_id"`
	ServerID  uint `json:"server_id"`
}

func NewSwarmInitTask(clusterID, serverID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(SwarmInitPayload{ClusterID: clusterID, ServerID: serverID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSwarmInit, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

func NewSwarmJoinTask(clusterID, serverID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(SwarmJoinPayload{ClusterID: clusterID, ServerID: serverID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSwarmJoin, payload, asynq.Queue("provisioning"), asynq.MaxRetry(3)), nil
}

// SwarmTaskHandler handles Docker Swarm provisioning tasks.
type SwarmTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

// HandleSwarmInit initializes Docker Swarm on the manager node.
func (h *SwarmTaskHandler) HandleSwarmInit(ctx context.Context, t *asynq.Task) error {
	var payload SwarmInitPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	log.Printf("Initializing Docker Swarm on server %d for cluster %d", payload.ServerID, payload.ClusterID)

	var cluster model.Cluster
	if err := h.DB.First(&cluster, payload.ClusterID).Error; err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	var server model.Server
	if err := h.DB.First(&server, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	h.DB.Model(&cluster).Update("status", model.ClusterStatusProvisioning)

	sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		h.setClusterError(&cluster, "failed to decrypt SSH key")
		return fmt.Errorf("decrypt SSH key: %w", err)
	}

	client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
	if err != nil {
		h.setClusterError(&cluster, fmt.Sprintf("SSH failed: %v", err))
		return fmt.Errorf("SSH failed: %w", err)
	}
	defer client.Close()

	// Install Docker if not present
	installCmd := `command -v docker >/dev/null 2>&1 || { curl -fsSL https://get.docker.com | sh; }`
	if result, err := client.ExecuteCommand(installCmd); err != nil {
		h.setClusterError(&cluster, fmt.Sprintf("Docker install failed: %s", result.Stderr))
		return fmt.Errorf("Docker install failed: %w", err)
	}

	// Initialize Swarm
	initCmd := fmt.Sprintf("docker swarm init --advertise-addr %s 2>/dev/null || echo ALREADY_SWARM", server.IP)
	result, err := client.ExecuteCommand(initCmd)
	if err != nil {
		h.setClusterError(&cluster, fmt.Sprintf("Swarm init failed: %s", result.Stderr))
		return fmt.Errorf("swarm init failed: %w", err)
	}

	// Get worker join token
	tokenResult, err := client.ExecuteCommand("docker swarm join-token worker -q")
	if err != nil {
		h.setClusterError(&cluster, "failed to get swarm join token")
		return fmt.Errorf("get join token: %w", err)
	}
	joinToken := strings.TrimSpace(tokenResult.Stdout)

	// Update cluster
	h.DB.Model(&cluster).Updates(map[string]interface{}{
		"swarm_join_token": joinToken,
		"status":           model.ClusterStatusActive,
	})

	h.DB.Model(&server).Updates(map[string]interface{}{
		"role":       model.ServerRoleManager,
		"cluster_id": cluster.ID,
	})

	log.Printf("Docker Swarm initialized for cluster %d", payload.ClusterID)
	return nil
}

// HandleSwarmJoin joins a worker to the Docker Swarm cluster.
func (h *SwarmTaskHandler) HandleSwarmJoin(ctx context.Context, t *asynq.Task) error {
	var payload SwarmJoinPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	log.Printf("Joining server %d to Swarm cluster %d", payload.ServerID, payload.ClusterID)

	var cluster model.Cluster
	if err := h.DB.Preload("ManagerServer").First(&cluster, payload.ClusterID).Error; err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	if cluster.SwarmJoinToken == "" {
		return fmt.Errorf("cluster %d has no swarm join token, manager not ready", payload.ClusterID)
	}

	var worker model.Server
	if err := h.DB.First(&worker, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	sshKey, err := decrypt(worker.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		return fmt.Errorf("decrypt SSH key: %w", err)
	}

	client, err := sshpkg.NewClient(worker.IP, worker.SSHPort, worker.SSHUser, sshKey, "")
	if err != nil {
		return fmt.Errorf("SSH failed: %w", err)
	}
	defer client.Close()

	// Install Docker if not present
	installCmd := `command -v docker >/dev/null 2>&1 || { curl -fsSL https://get.docker.com | sh; }`
	if result, err := client.ExecuteCommand(installCmd); err != nil {
		return fmt.Errorf("Docker install failed: %s %w", result.Stderr, err)
	}

	// Join Swarm
	joinCmd := fmt.Sprintf("docker swarm join --token %s %s:2377",
		cluster.SwarmJoinToken, cluster.ManagerServer.IP)
	result, err := client.ExecuteCommand(joinCmd)
	if err != nil {
		return fmt.Errorf("swarm join failed: %v\n%s", err, result.Stderr)
	}

	h.DB.Model(&worker).Updates(map[string]interface{}{
		"role":       model.ServerRoleWorker,
		"cluster_id": cluster.ID,
	})

	log.Printf("Server %d joined Swarm cluster %d", payload.ServerID, payload.ClusterID)
	return nil
}

func (h *SwarmTaskHandler) setClusterError(cluster *model.Cluster, msg string) {
	h.DB.Model(cluster).Updates(map[string]interface{}{
		"status":        model.ClusterStatusError,
		"error_message": msg,
	})
}
