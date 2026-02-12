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

const TypePushEnv = "env:push"

type PushEnvPayload struct {
	EnvironmentID uint `json:"environment_id"`
}

func NewPushEnvTask(envID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(PushEnvPayload{EnvironmentID: envID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypePushEnv, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

type EnvTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

// HandlePushEnv pushes environment variables to all servers in a cluster.
func (h *EnvTaskHandler) HandlePushEnv(ctx context.Context, t *asynq.Task) error {
	var payload PushEnvPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var env models.Environment
	if err := h.DB.Preload("Cluster").First(&env, payload.EnvironmentID).Error; err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	// Get all servers in this cluster
	var servers []models.Server
	if err := h.DB.Where("cluster_id = ?", env.ClusterID).Find(&servers).Error; err != nil {
		return fmt.Errorf("fetch servers: %w", err)
	}

	if len(servers) == 0 {
		log.Printf("No servers in cluster %d for env push", env.ClusterID)
		h.DB.Model(&env).Update("synced", true)
		return nil
	}

	// Build .env file content
	var envLines []string
	for k, v := range env.Variables {
		envLines = append(envLines, fmt.Sprintf("%s=%s", k, v))
	}
	envContent := strings.Join(envLines, "\n")

	envDir := "/opt/orchestra/envs"
	envFile := fmt.Sprintf("%s/%s-%s.env", envDir, sanitizeName(env.Cluster.Name), string(env.Scope))

	for _, server := range servers {
		sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
		if err != nil {
			log.Printf("Failed to decrypt SSH key for server %d: %v", server.ID, err)
			continue
		}

		client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
		if err != nil {
			log.Printf("SSH failed for server %d: %v", server.ID, err)
			continue
		}

		client.ExecuteCommand(fmt.Sprintf("mkdir -p %s", envDir))
		writeCmd := fmt.Sprintf("cat > %s << 'ORCHESTRA_EOF'\n%s\nORCHESTRA_EOF", envFile, envContent)
		if result, err := client.ExecuteCommand(writeCmd); err != nil {
			log.Printf("Failed to push env to server %d: %s", server.ID, result.Stderr)
		} else {
			log.Printf("Pushed env to server %d: %s", server.ID, envFile)
		}

		client.Close()
	}

	h.DB.Model(&env).Update("synced", true)
	log.Printf("Environment %d pushed to %d servers", env.ID, len(servers))
	return nil
}
