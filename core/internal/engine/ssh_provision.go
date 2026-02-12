package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/enochcodes/orchestra/core/internal/model"
	sshpkg "github.com/enochcodes/orchestra/core/pkg/ssh"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const (
	// TypePreflightCheck is the Asynq task type for pre-flight server checks.
	TypePreflightCheck = "server:preflight_check"

	// TypeInstallK3s is the Asynq task type for installing K3s on a server.
	TypeInstallK3s = "server:install_k3s"
)

// PreflightPayload contains the data needed to run a pre-flight check.
type PreflightPayload struct {
	ServerID uint `json:"server_id"`
}

// InstallK3sPayload contains the data for K3s installation.
type InstallK3sPayload struct {
	ServerID  uint   `json:"server_id"`
	Role      string `json:"role"` // "server" or "agent"
	Token     string `json:"token,omitempty"`
	ServerURL string `json:"server_url,omitempty"`
}

// NewPreflightCheckTask creates a new Asynq task for pre-flight checking a server.
func NewPreflightCheckTask(serverID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(PreflightPayload{ServerID: serverID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preflight payload: %w", err)
	}
	return asynq.NewTask(TypePreflightCheck, payload, asynq.Queue("provisioning"), asynq.MaxRetry(3)), nil
}

// NewInstallK3sTask creates a new Asynq task for installing K3s.
func NewInstallK3sTask(serverID uint, role, token, serverURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(InstallK3sPayload{
		ServerID:  serverID,
		Role:      role,
		Token:     token,
		ServerURL: serverURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal install payload: %w", err)
	}
	return asynq.NewTask(TypeInstallK3s, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

// SSHProvisionHandler handles SSH-based provisioning tasks.
type SSHProvisionHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

// HandlePreflightCheck connects to a server via SSH and runs pre-flight checks.
func (h *SSHProvisionHandler) HandlePreflightCheck(ctx context.Context, t *asynq.Task) error {
	var payload PreflightPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting pre-flight check for server ID: %d", payload.ServerID)

	// Fetch server details from DB
	var server model.Server
	if err := h.DB.First(&server, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Update status to preflight
	h.DB.Model(&server).Update("status", model.ServerStatusPreflight)

	// Decrypt SSH key
	sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		h.setServerError(&server, "failed to decrypt SSH key")
		return fmt.Errorf("failed to decrypt SSH key: %w", err)
	}

	// Connect via SSH
	client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
	if err != nil {
		h.setServerError(&server, fmt.Sprintf("SSH connection failed: %v", err))
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Run pre-flight checks
	report, err := sshpkg.RunPreflightCheck(client)
	if err != nil {
		h.setServerError(&server, fmt.Sprintf("preflight check failed: %v", err))
		return fmt.Errorf("preflight check failed: %w", err)
	}

	// Update server with results
	status := model.ServerStatusReady
	if !report.Compatible {
		status = model.ServerStatusError
	}

	h.DB.Model(&server).Updates(map[string]interface{}{
		"status":           status,
		"os":               report.OS,
		"arch":             report.Arch,
		"cpu_cores":        report.CPUCores,
		"ram_bytes":        report.RAMBytes,
		"preflight_report": report.ToJSON(),
	})

	log.Printf("Pre-flight check completed for server %d: compatible=%v", payload.ServerID, report.Compatible)
	return nil
}

// HandleInstallK3s installs K3s on a remote server via SSH.
func (h *SSHProvisionHandler) HandleInstallK3s(ctx context.Context, t *asynq.Task) error {
	var payload InstallK3sPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting K3s installation on server ID: %d (role: %s)", payload.ServerID, payload.Role)

	// Fetch server details
	var server model.Server
	if err := h.DB.First(&server, payload.ServerID).Error; err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Decrypt SSH key
	sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt SSH key: %w", err)
	}

	// Connect via SSH
	client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Build the K3s install command
	var installCmd string
	if payload.Role == "server" {
		installCmd = "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='server' sh -"
	} else {
		installCmd = fmt.Sprintf(
			"curl -sfL https://get.k3s.io | K3S_URL='%s' K3S_TOKEN='%s' sh -",
			payload.ServerURL,
			payload.Token,
		)
	}

	// Execute K3s installation
	result, err := client.ExecuteCommand(installCmd)
	if err != nil {
		h.setServerError(&server, fmt.Sprintf("K3s installation failed: %v\nStderr: %s", err, result.Stderr))
		return fmt.Errorf("K3s installation failed: %w", err)
	}

	// Update server role
	role := model.ServerRoleManager
	if payload.Role == "agent" {
		role = model.ServerRoleWorker
	}
	h.DB.Model(&server).Update("role", role)

	log.Printf("K3s installation completed on server %d", payload.ServerID)
	return nil
}

// setServerError updates a server's status to error with a message.
func (h *SSHProvisionHandler) setServerError(server *model.Server, msg string) {
	h.DB.Model(server).Updates(map[string]interface{}{
		"status":        model.ServerStatusError,
		"error_message": msg,
	})
}
