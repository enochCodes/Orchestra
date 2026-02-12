package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SystemRole defines the global platform role.
type SystemRole string

const (
	SystemRoleAdmin     SystemRole = "system_admin"
	SystemRoleDeveloper SystemRole = "developer"
)

// ServerRoleType defines the role a user has on a specific server.
type ServerRoleType string

const (
	ServerRoleTypeAdmin  ServerRoleType = "admin"
	ServerRoleTypeDevOps ServerRoleType = "devops"
	ServerRoleTypeViewer ServerRoleType = "viewer"
)

// User represents a platform user.
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	DisplayName  string         `gorm:"size:255" json:"display_name"`
	Avatar       string         `gorm:"size:500" json:"avatar,omitempty"`
	SystemRole   SystemRole     `gorm:"size:30;default:'developer'" json:"system_role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}

// SetPassword hashes and stores the password.
func (u *User) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password.
func (u *User) CheckPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain))
	return err == nil
}

// IsSystemAdmin returns true if user is system admin.
func (u *User) IsSystemAdmin() bool {
	return u.SystemRole == SystemRoleAdmin
}

// ServerMembership links a user to a server with a role.
type ServerMembership struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_server_user" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ServerID  uint           `gorm:"not null;uniqueIndex:idx_server_user" json:"server_id"`
	Server    Server         `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Role      ServerRoleType `gorm:"size:20;default:'viewer'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (ServerMembership) TableName() string {
	return "server_memberships"
}

