package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/enochcodes/orchestra/core/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const maxRetries = 15
const retryDelay = 3 * time.Second

// Connect establishes a connection to PostgreSQL and runs auto-migrations.
func Connect(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

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

		log.Println("Connected to PostgreSQL database")
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Fix columns that AutoMigrate won't alter (jsonb -> text for empty-string compat).
	// These are idempotent — if already text, PostgreSQL returns a no-op.
	fixes := []string{
		`DO $$ BEGIN ALTER TABLE servers ALTER COLUMN preflight_report TYPE text USING preflight_report::text; EXCEPTION WHEN undefined_table THEN NULL; END $$;`,
		`DO $$ BEGIN ALTER TABLE activities ALTER COLUMN metadata TYPE text USING metadata::text; EXCEPTION WHEN undefined_table THEN NULL; END $$;`,
	}
	for _, fix := range fixes {
		if err := db.Exec(fix).Error; err != nil {
			log.Printf("Column fix warning: %v", err)
		}
	}

	// Migrate in dependency order. constraint:false on circular refs avoids FK issues.
	modelsToMigrate := []interface{}{
		&model.User{},
		&model.ServerTeam{},
		&model.Server{},
		&model.ServerMembership{},
		&model.Cluster{},
		&model.Application{},
		&model.ApplicationMembership{},
		&model.Deployment{},
		&model.Activity{},
		&model.Environment{},
		&model.NginxConfig{},
	}
	for _, m := range modelsToMigrate {
		if err := db.AutoMigrate(m); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate %T: %w", m, err)
		}
	}

	log.Println("Database migrations completed")
	return db, nil
}
