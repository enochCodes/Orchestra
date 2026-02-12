package store

import (
	"log"

	"github.com/enochcodes/orchestra/core/internal/model"
	"gorm.io/gorm"
)

// Seed creates initial data (system admin user).
func Seed(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping seed")
		return nil
	}

	admin := model.User{
		Email:       "admin@orchestra.local",
		DisplayName: "System Admin",
		SystemRole:  model.SystemRoleAdmin,
	}
	if err := admin.SetPassword("admin123"); err != nil {
		return err
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	log.Printf("Created system admin user: %s (password: admin123)", admin.Email)
	return nil
}
