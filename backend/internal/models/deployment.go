package models

import (
	"time"

	"gorm.io/gorm"
)

// DeploymentStatus represents the lifecycle state of a deployment.
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusBuilding  DeploymentStatus = "building"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusLive      DeploymentStatus = "live"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
)

// Deployment represents a single deployment attempt for an application.
type Deployment struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	ApplicationID uint             `gorm:"not null" json:"application_id"`
	Application   Application      `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Version       string           `gorm:"size:100;not null" json:"version"`
	ImageTag      string           `gorm:"size:255" json:"image_tag"`
	Status        DeploymentStatus `gorm:"size:20;default:'pending'" json:"status"`
	Logs          string           `gorm:"type:text" json:"logs,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (Deployment) TableName() string {
	return "deployments"
}
