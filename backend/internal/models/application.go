package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// DeploymentSourceType defines how the application is deployed.
type DeploymentSourceType string

const (
	DeploymentSourceGit    DeploymentSourceType = "git"
	DeploymentSourceManual DeploymentSourceType = "manual"
	DeploymentSourceDocker DeploymentSourceType = "docker_image"
)

// Application represents a deployable application bound to a cluster.
type Application struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	Name      string  `gorm:"size:255;not null" json:"name"`
	ClusterID uint    `gorm:"not null" json:"cluster_id"`
	Cluster   Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	Namespace string  `gorm:"size:255;not null" json:"namespace"`

	// Deployment Source
	SourceType   DeploymentSourceType `gorm:"size:20;default:'git'" json:"source_type"`
	RepoURL      string               `gorm:"size:500" json:"repo_url"`
	Branch       string               `gorm:"size:100;default:'main'" json:"branch"`
	DockerImage  string               `gorm:"size:500" json:"docker_image"`  // for docker_image source
	ManualPath   string               `gorm:"size:500" json:"manual_path"`  // for manual upload path

	// Build Configuration
	BuildType string     `gorm:"size:50;default:'docker'" json:"build_type"` // go, node, python, docker, nextjs-static
	BuildCmd  string     `gorm:"size:255" json:"build_cmd"`
	StartCmd  string     `gorm:"size:255" json:"start_cmd"`
	EnvVars   ScopedEnvs `gorm:"type:jsonb" json:"env_vars"`

	Replicas  int            `gorm:"default:1" json:"replicas"`
	Status    string         `gorm:"size:20;default:'pending'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ApplicationRole defines the role a user has on an application.
type ApplicationRole string

const (
	ApplicationRoleAdmin   ApplicationRole = "admin"
	ApplicationRoleManager ApplicationRole = "manager"
	ApplicationRoleViewer  ApplicationRole = "viewer"
)

// ApplicationMembership links a user to an application with a role.
type ApplicationMembership struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	UserID        uint            `gorm:"not null;uniqueIndex:idx_app_user" json:"user_id"`
	User          User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ApplicationID uint            `gorm:"not null;uniqueIndex:idx_app_user" json:"application_id"`
	Application   Application     `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Role          ApplicationRole `gorm:"size:20;default:'viewer'" json:"role"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (ApplicationMembership) TableName() string {
	return "application_memberships"
}

type ScopedEnvs struct {
	Production map[string]string `json:"production"`
	Preview    map[string]string `json:"preview"`
}

// Value Marshal
func (m ScopedEnvs) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan Unmarshal
func (m *ScopedEnvs) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &m)
}

// TableName overrides the table name.
func (Application) TableName() string {
	return "applications"
}
