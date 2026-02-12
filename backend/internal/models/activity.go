package models

import (
	"time"

	"gorm.io/gorm"
)

// ActivityType represents the type of activity.
type ActivityType string

const (
	ActivityTypeServerRegistered ActivityType = "server_registered"
	ActivityTypeClusterCreated   ActivityType = "cluster_created"
	ActivityTypeClusterProvisioned ActivityType = "cluster_provisioned"
	ActivityTypeAppDeployed      ActivityType = "app_deployed"
	ActivityTypeDeploymentFailed ActivityType = "deployment_failed"
	ActivityTypeUserLogin        ActivityType = "user_login"
)

// Activity represents an audit/activity log entry.
type Activity struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	Type      ActivityType `gorm:"size:50;not null" json:"type"`
	Message   string       `gorm:"type:text;not null" json:"message"`
	Entity    string       `gorm:"size:50" json:"entity"` // server, cluster, application, deployment
	EntityID  uint         `json:"entity_id"`
	UserID    *uint         `json:"user_id,omitempty"`
	Metadata  string       `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (Activity) TableName() string {
	return "activities"
}
