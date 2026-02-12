package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/enochcodes/orchestra/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const maxRetries = 15
const retryDelay = 3 * time.Second

// Connect establishes a connection to PostgreSQL and runs auto-migrations.
// Retries up to 15 times with 3s delay for Docker startup.
func Connect(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger:                           logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true, // Resolves Cluster<->Server circular FK + concurrent migration races
		})
		if err != nil {
			log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		// Verify connection with ping
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Database connection attempt %d/%d: failed to get DB: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := sqlDB.PingContext(ctx); err != nil {
			cancel()
			log.Printf("Database connection attempt %d/%d: ping failed: %v", i+1, maxRetries, err)
			sqlDB.Close()
			time.Sleep(retryDelay)
			continue
		}
		cancel()

		// Connection OK
		log.Println("Connected to PostgreSQL database")
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Auto-migrate models in explicit order. Server before Cluster: Cluster has Workers []Server
	// (foreignKey:ClusterID), so GORM may alter servers table when migrating Cluster.
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.ServerTeam{},
		&models.Server{},
		&models.ServerMembership{},
		&models.Cluster{},
		&models.Application{},
		&models.ApplicationMembership{},
		&models.Deployment{},
		&models.Activity{},
	}
	for _, m := range modelsToMigrate {
		if err := db.AutoMigrate(m); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate %T: %w", m, err)
		}
	}

	log.Println("Database migrations completed")
	return db, nil
}
