package services

import (
	"fmt"
	"log"
	"server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func PostgresDB() *gorm.DB {
	return _db
}

func LoadDatabase() {
	Logger.Info("Connecting to database")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		Conf.DB.HOST,
		Conf.DB.USER,
		Conf.DB.PASS,
		Conf.DB.NAME,
		Conf.DB.PORT,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	_db = database

	loadMigrations(_db)
}

func loadMigrations(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal(err)
	}
}
