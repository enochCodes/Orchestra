package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/enochcodes/orchestra/backend/internal/models"
	sshpkg "github.com/enochcodes/orchestra/backend/pkg/ssh"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const (
	// TypeDesignateManager is the Asynq task type for setting up a K3s server (manager).
	TypeDesignateManager = "cluster:designate_manager"

	// TypeJoinWorker is the Asynq task type for joining a worker to a cluster.
	TypeJoinWorker = "cluster:join_worker"
)

// DesignateManagerPayload contains the data for manager setup.
type DesignateManagerPayload struct {
	ClusterID uint `json:"cluster_id"`
	ServerID  uint `json:"server_id"`
}

// JoinWorkerPayload contains the data for worker join.
type JoinWorkerPayload struct {
	ClusterID uint `json:"cluster_id"`
	ServerID  uint `json:"server_id"`
}

// NewDesignateManagerTask creates a task to install K3s server on the manager node.
func NewDesignateManagerTask(clusterID, serverID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(DesignateManagerPayload{
		ClusterID: clusterID,
		ServerID:  serverID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeDesignateManager, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

// NewJoinWorkerTask creates a task to join a worker node to the cluster.
func NewJoinWorkerTask(clusterID, serverID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(JoinWorkerPayload{
		ClusterID: clusterID,
		ServerID:  serverID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeJoinWorker, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

// K8sTaskHandler handles Kubernetes cluster provisioning tasks.
type K8sTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

// HandleDesignateManager installs K3s server on the manager node and retrieves the kubeconfig + node token.
func (h *K8sTaskHandler) HandleDesignateManager(ctx context.Context, t *asynq.Task) error {
	var payload DesignateManagerPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Designating server %d as manager for cluster %d", payload.ServerID, payload.ClusterID)

	// Fetch cluster and server
	var cluster models.Cluster
	if err := h.DB.First(&cluster, payload.ClusterID).Error; err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	var server models.Server
	if err := h.DB.First(&server, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Update cluster status
	h.DB.Model(&cluster).Update("status", models.ClusterStatusProvisioning)

	// Decrypt SSH key
	sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		h.setClusterError(&cluster, "failed to decrypt SSH key")
		return fmt.Errorf("failed to decrypt SSH key: %w", err)
	}

	// Connect via SSH
	client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
	if err != nil {
		h.setClusterError(&cluster, fmt.Sprintf("SSH failed: %v", err))
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Install K3s server
	result, err := client.ExecuteCommand("curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='server' sh -")
	if err != nil {
		h.setClusterError(&cluster, fmt.Sprintf("K3s server install failed: %s", result.Stderr))
		return fmt.Errorf("K3s install failed: %w", err)
	}

	// Wait a moment then retrieve kubeconfig
	result, err = client.ExecuteCommand("cat /etc/rancher/k3s/k3s.yaml")
	if err != nil {
		h.setClusterError(&cluster, "failed to retrieve kubeconfig")
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	kubeconfig := result.Stdout

	// Replace localhost with the server's actual IP in the kubeconfig
	kubeconfig = strings.ReplaceAll(kubeconfig, "127.0.0.1", server.IP)
	kubeconfig = strings.ReplaceAll(kubeconfig, "localhost", server.IP)

	// Retrieve the node token for joining workers
	result, err = client.ExecuteCommand("cat /var/lib/rancher/k3s/server/node-token")
	if err != nil {
		h.setClusterError(&cluster, "failed to retrieve node token")
		return fmt.Errorf("failed to get node token: %w", err)
	}
	nodeToken := strings.TrimSpace(result.Stdout)

	// Encrypt and store kubeconfig
	encryptedKubeconfig, err := encrypt([]byte(kubeconfig), h.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt kubeconfig: %w", err)
	}

	// Update cluster with kubeconfig and token
	h.DB.Model(&cluster).Updates(map[string]interface{}{
		"kubeconfig_encrypted": encryptedKubeconfig,
		"node_token":           nodeToken,
		"status":               models.ClusterStatusActive,
	})

	// Update server role
	h.DB.Model(&server).Updates(map[string]interface{}{
		"role":       models.ServerRoleManager,
		"cluster_id": cluster.ID,
	})

	log.Printf("Manager designation completed for cluster %d", payload.ClusterID)
	return nil
}

// HandleJoinWorker joins a worker node to an existing cluster using the node token.
func (h *K8sTaskHandler) HandleJoinWorker(ctx context.Context, t *asynq.Task) error {
	var payload JoinWorkerPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Joining server %d as worker to cluster %d", payload.ServerID, payload.ClusterID)

	// Fetch cluster with manager details
	var cluster models.Cluster
	if err := h.DB.Preload("ManagerServer").First(&cluster, payload.ClusterID).Error; err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	if cluster.NodeToken == "" {
		return fmt.Errorf("cluster %d has no node token, manager not ready", payload.ClusterID)
	}

	// Fetch worker server
	var worker models.Server
	if err := h.DB.First(&worker, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Decrypt worker SSH key
	sshKey, err := decrypt(worker.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt SSH key: %w", err)
	}

	// Connect via SSH to the worker
	client, err := sshpkg.NewClient(worker.IP, worker.SSHPort, worker.SSHUser, sshKey, "")
	if err != nil {
		return fmt.Errorf("SSH connection to worker failed: %w", err)
	}
	defer client.Close()

	// Join the worker to the cluster
	managerURL := fmt.Sprintf("https://%s:6443", cluster.ManagerServer.IP)
	joinCmd := fmt.Sprintf(
		"curl -sfL https://get.k3s.io | K3S_URL='%s' K3S_TOKEN='%s' sh -",
		managerURL,
		cluster.NodeToken,
	)

	result, err := client.ExecuteCommand(joinCmd)
	if err != nil {
		return fmt.Errorf("K3s agent join failed: %v\nStderr: %s", err, result.Stderr)
	}

	// Update worker record
	h.DB.Model(&worker).Updates(map[string]interface{}{
		"role":       models.ServerRoleWorker,
		"cluster_id": cluster.ID,
	})

	log.Printf("Worker %d joined cluster %d successfully", payload.ServerID, payload.ClusterID)
	return nil
}

// setClusterError updates a cluster's status to error with a message.
func (h *K8sTaskHandler) setClusterError(cluster *models.Cluster, msg string) {
	h.DB.Model(cluster).Updates(map[string]interface{}{
		"status":        models.ClusterStatusError,
		"error_message": msg,
	})
}
