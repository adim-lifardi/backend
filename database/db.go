package database

import (
	"finance/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// Ambil env dengan fallback default
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}

	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	if dbname == "" {
		dbname = "postgres"
	}

	port := os.Getenv("DATABASE_PORT")
	if port == "" {
		port = "5432"
	}

	// DSN format untuk Postgres
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Asia/Jakarta",
		host, user, password, dbname, port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
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

	log.Println("Postgres connected & migrated successfully!")
}
