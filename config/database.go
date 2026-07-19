package config

import (
	"log"
	"os"

	"github.com/raihansyahrin/backend_laundry_app.git/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(localhost:3308)/backend_laundry_app?charset=utf8mb4&parseTime=True&loc=Local"
	}

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		panic(err)
	}

	err = database.AutoMigrate(&models.User{}, &models.Address{}, &models.Order{}, &models.Service{}, &models.PasswordResetOTP{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = database

	// Seed database with dummy data
	SeedDatabase()
}
