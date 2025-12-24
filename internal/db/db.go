package db

import (
	"fmt"
	"log"
	"os"
	"warehouse-backend/internal/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	_ = godotenv.Load()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to DB: ", err)
	}
	log.Println("✅ Connected to database")
	return db
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Item{},
		&model.ItemImage{},
		&model.Sale{},
		&model.User{},
	)
	if err != nil {
		log.Fatal("❌ Migration error: ", err)
	}
	log.Println("✅ Database migrated")
}
