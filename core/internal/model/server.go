package model

import (
	"time"

	"gorm.io/gorm"
)

// ServerStatus represents the lifecycle state of a server.
type ServerStatus string

const (
	ServerStatusPending   ServerStatus = "pending"
	ServerStatusPreflight ServerStatus = "preflight"
	ServerStatusReady     ServerStatus = "ready"
	ServerStatusError     ServerStatus = "error"
)

// ServerRole defines the Kubernetes role assigned to a server.
type ServerRole string

const (
	ServerRoleNone    ServerRole = "none"
	ServerRoleManager ServerRole = "manager"
	ServerRoleWorker  ServerRole = "worker"
)

// Server represents a physical server registered in the inventory.
type Server struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Hostname        string         `gorm:"size:255" json:"hostname"`
	IP              string         `gorm:"size:45;not null;uniqueIndex" json:"ip"`
	SSHPort         int            `gorm:"default:22" json:"ssh_port"`
	SSHUser         string         `gorm:"size:255;not null" json:"ssh_user"`
	SSHKeyEncrypted []byte         `gorm:"type:bytea" json:"-"`
	OS              string         `gorm:"size:100" json:"os"`
	Arch            string         `gorm:"size:50" json:"arch"`
	CPUCores        int            `json:"cpu_cores"`
	RAMBytes        int64          `json:"ram_bytes"`
	DiskInfo        string         `gorm:"type:text" json:"disk_info"`
	Status          ServerStatus   `gorm:"size:20;default:'pending'" json:"status"`
	Role            ServerRole     `gorm:"size:20;default:'none'" json:"role"`
	PreflightReport string         `gorm:"type:text" json:"preflight_report,omitempty"`
	ClusterID       *uint          `json:"cluster_id,omitempty"`
	Cluster         *Cluster       `gorm:"foreignKey:ClusterID;constraint:false" json:"cluster,omitempty"`
	TeamID          *uint          `json:"team_id,omitempty"`
	Team            *ServerTeam    `gorm:"foreignKey:TeamID;constraint:false" json:"team,omitempty"`
	CreatedByUserID *uint          `json:"created_by_user_id,omitempty"`
	ErrorMessage    string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (Server) TableName() string {
	return "servers"
}

// ServerTeam represents a team for grouping servers.
type ServerTeam struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (ServerTeam) TableName() string {
	return "server_teams"
}
