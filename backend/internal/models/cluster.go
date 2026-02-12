package models

import (
	"time"

	"gorm.io/gorm"
)

// ClusterStatus represents the lifecycle state of a cluster.
type ClusterStatus string

const (
	ClusterStatusPending      ClusterStatus = "pending"
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusActive       ClusterStatus = "active"
	ClusterStatusDegraded     ClusterStatus = "degraded"
	ClusterStatusError        ClusterStatus = "error"
)

// Cluster represents a Kubernetes cluster composed of physical servers.
type Cluster struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Name                string         `gorm:"size:255;not null;uniqueIndex" json:"name"`
	ManagerServerID     uint           `gorm:"not null" json:"manager_server_id"`
	ManagerServer       Server         `gorm:"foreignKey:ManagerServerID;constraint:false" json:"manager_server,omitempty"`
	Workers             []Server       `gorm:"foreignKey:ClusterID;constraint:false" json:"workers,omitempty"`
	KubeconfigEncrypted []byte         `gorm:"type:bytea" json:"-"`
	NodeToken           string         `gorm:"type:text" json:"-"`
	CNIPlugin           string         `gorm:"column:cni_plugin;size:50;default:'flannel'" json:"cni_plugin"`
	Status              ClusterStatus  `gorm:"size:20;default:'pending'" json:"status"`
	ErrorMessage        string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (Cluster) TableName() string {
	return "clusters"
}
