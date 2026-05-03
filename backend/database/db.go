package database

import (
	"os"
	"strings"

	"github.com/ahmadfarisfs/bajet/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(path string) error {
	var err error
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	dbURL := os.Getenv("DATABASE_URL")
	if strings.HasPrefix(dbURL, "postgres") {
		DB, err = gorm.Open(postgres.Open(dbURL), cfg)
	} else {
		DB, err = gorm.Open(sqlite.Open(path), cfg)
	}
	if err != nil {
		return err
	}
	return DB.AutoMigrate(&models.Cycle{}, &models.Period{}, &models.Waitlist{})
}
