package database

import (
	"github.com/ahmadfarisfs/bajet/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(path string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}
	return DB.AutoMigrate(&models.Cycle{}, &models.Period{})
}
