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

const TypeNginxProvision = "server:nginx_provision"

type NginxProvisionPayload struct {
	NginxConfigID uint `json:"nginx_config_id"`
}

func NewNginxProvisionTask(configID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(NginxProvisionPayload{NginxConfigID: configID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeNginxProvision, payload, asynq.Queue("provisioning"), asynq.MaxRetry(2)), nil
}

type NginxTaskHandler struct {
	DB            *gorm.DB
	EncryptionKey string
}

func (h *NginxTaskHandler) HandleNginxProvision(ctx context.Context, t *asynq.Task) error {
	var payload NginxProvisionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var cfg models.NginxConfig
	if err := h.DB.Preload("Server").First(&cfg, payload.NginxConfigID).Error; err != nil {
		return fmt.Errorf("nginx config not found: %w", err)
	}

	server := cfg.Server
	sshKey, err := decrypt(server.SSHKeyEncrypted, h.EncryptionKey)
	if err != nil {
		h.setStatus(&cfg, "error")
		return fmt.Errorf("decrypt SSH key: %w", err)
	}

	client, err := sshpkg.NewClient(server.IP, server.SSHPort, server.SSHUser, sshKey, "")
	if err != nil {
		h.setStatus(&cfg, "error")
		return fmt.Errorf("SSH failed: %w", err)
	}
	defer client.Close()

	// Install nginx if not present
	installCmd := `command -v nginx >/dev/null 2>&1 || { apt-get update -qq && apt-get install -y -qq nginx; } || { yum install -y nginx; }`
	client.ExecuteCommand(installCmd)

	// Generate nginx config
	nginxConf := h.generateNginxConfig(&cfg)

	// Write config
	confPath := fmt.Sprintf("/etc/nginx/sites-available/%s", sanitizeName(cfg.Domain))
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", sanitizeName(cfg.Domain))

	writeCmd := fmt.Sprintf("cat > %s << 'ORCHESTRA_EOF'\n%s\nORCHESTRA_EOF", confPath, nginxConf)
	if result, err := client.ExecuteCommand(writeCmd); err != nil {
		h.setStatus(&cfg, "error")
		return fmt.Errorf("write nginx config: %s %w", result.Stderr, err)
	}

	// Enable site
	client.ExecuteCommand("mkdir -p /etc/nginx/sites-enabled")
	client.ExecuteCommand(fmt.Sprintf("ln -sf %s %s", confPath, enabledPath))

	// Test and reload nginx
	result, err := client.ExecuteCommand("nginx -t 2>&1 && systemctl reload nginx 2>&1")
	if err != nil {
		h.setStatus(&cfg, "error")
		return fmt.Errorf("nginx reload failed: %s %w", result.Stderr, err)
	}

	// Setup Let's Encrypt if requested
	if cfg.LetsEncrypt && cfg.SSLEnabled {
		log.Printf("Setting up Let's Encrypt for %s", cfg.Domain)
		certCmd := fmt.Sprintf(
			`command -v certbot >/dev/null 2>&1 || { apt-get install -y -qq certbot python3-certbot-nginx; } && certbot --nginx -d %s --non-interactive --agree-tos --email admin@%s 2>&1`,
			cfg.Domain, cfg.Domain,
		)
		client.ExecuteCommand(certCmd)
	}

	h.setStatus(&cfg, "active")
	log.Printf("Nginx configured for %s on server %d", cfg.Domain, server.ID)
	return nil
}

func (h *NginxTaskHandler) generateNginxConfig(cfg *models.NginxConfig) string {
	if cfg.CustomConfig != "" {
		return cfg.CustomConfig
	}

	upstream := fmt.Sprintf("http://127.0.0.1:%d", cfg.UpstreamPort)

	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass %s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}`, cfg.Domain, upstream)
}

func (h *NginxTaskHandler) setStatus(cfg *models.NginxConfig, status string) {
	h.DB.Model(cfg).Update("status", status)
}
