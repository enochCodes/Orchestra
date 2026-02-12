package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// EnvScope differentiates production vs preview environments.
type EnvScope string

const (
	EnvScopeProduction EnvScope = "production"
	EnvScopePreview    EnvScope = "preview"
	EnvScopeStaging    EnvScope = "staging"
)

// Environment stores a set of key-value environment variables for a cluster+scope.
type Environment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ClusterID uint           `gorm:"not null" json:"cluster_id"`
	Cluster   Cluster        `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	Scope     EnvScope       `gorm:"size:30;not null;default:'production'" json:"scope"`
	Name      string         `gorm:"size:255;not null" json:"name"` // e.g. "production-v1", "staging"
	Variables EnvVarMap      `gorm:"type:jsonb" json:"variables"`
	Synced    bool           `gorm:"default:false" json:"synced"` // whether pushed to servers
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Environment) TableName() string {
	return "environments"
}

// EnvVarMap is a JSON-serializable map of environment variables.
type EnvVarMap map[string]string

func (m EnvVarMap) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(m)
}

func (m *EnvVarMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}

// NginxConfig stores nginx reverse-proxy settings for a server or cluster.
type NginxConfig struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ServerID      uint           `gorm:"not null" json:"server_id"`
	Server        Server         `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Domain        string         `gorm:"size:255;not null" json:"domain"`
	UpstreamPort  int            `gorm:"not null" json:"upstream_port"`
	SSLEnabled    bool           `gorm:"default:false" json:"ssl_enabled"`
	LetsEncrypt   bool           `gorm:"default:false" json:"lets_encrypt"`
	CustomConfig  string         `gorm:"type:text" json:"custom_config,omitempty"`
	ApplicationID *uint          `json:"application_id,omitempty"`
	Status        string         `gorm:"size:20;default:'pending'" json:"status"` // pending, active, error
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NginxConfig) TableName() string {
	return "nginx_configs"
}
