package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Password string
	Role     string `gorm:"default:'player'"`
}

func Init(dbType, dsn string) error {
	var dialector gorm.Dialector

	switch dbType {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Connected to %s database", dbType)

	// AutoMigrate
	if err := DB.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
