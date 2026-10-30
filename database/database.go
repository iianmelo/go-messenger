package database

import (
	"fmt"
	"log"
	"os"

	"go.messenger/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Dbinstance struct {
	Db *gorm.DB
}

var DB Dbinstance

func ConnectDb() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	log.Printf("connecting to postgres host=%s db=%s", os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}

	log.Println("connected")
	log.Println("running migrations")
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Chat{})
	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.GroupMember{})

	DB = Dbinstance{
		Db: db,
	}
	SeedIfEmpty()
}
