package database

import (
	"finance/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// Ambil konfigurasi dari environment variable
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Format DSN MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}

	// AutoMigrate semua model
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
		&models.Notification{},
		&models.Backup{},
	); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	log.Println("MySQL connected & migrated successfully!")
}
