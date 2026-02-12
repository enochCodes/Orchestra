package models

import (
	"time"

	"gorm.io/gorm"
)

// ClusterType represents the orchestration runtime.
type ClusterType string

const (
	ClusterTypeK8s         ClusterType = "k8s"
	ClusterTypeDockerSwarm ClusterType = "docker_swarm"
	ClusterTypeManual      ClusterType = "manual"
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

// Cluster represents a compute cluster composed of physical servers.
type Cluster struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Name                string         `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Type                ClusterType    `gorm:"column:cluster_type;size:30;default:'k8s'" json:"type"`
	ManagerServerID     uint           `gorm:"not null" json:"manager_server_id"`
	ManagerServer       Server         `gorm:"foreignKey:ManagerServerID;constraint:false" json:"manager_server,omitempty"`
	Workers             []Server       `gorm:"foreignKey:ClusterID;constraint:false" json:"workers,omitempty"`
	KubeconfigEncrypted []byte         `gorm:"type:bytea" json:"-"`
	NodeToken           string         `gorm:"type:text" json:"-"`
	SwarmJoinToken      string         `gorm:"type:text" json:"-"` // Docker Swarm worker join token
	CNIPlugin           string         `gorm:"column:cni_plugin;size:50;default:'flannel'" json:"cni_plugin"`
	Domain              string         `gorm:"size:255" json:"domain,omitempty"`
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
