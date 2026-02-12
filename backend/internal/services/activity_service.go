package services

import (
	"encoding/json"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"gorm.io/gorm"
)

// LogActivity creates an activity log entry. Returns error but typically we log and continue.
func LogActivity(db *gorm.DB, activityType models.ActivityType, message, entity string, entityID uint, userID *uint, metadata interface{}) error {
	var meta string
	if metadata != nil {
		b, _ := json.Marshal(metadata)
		meta = string(b)
	}
	a := models.Activity{
		Type:     activityType,
		Message:  message,
		Entity:   entity,
		EntityID: entityID,
		UserID:   userID,
		Metadata: meta,
	}
	return db.Create(&a).Error
}
