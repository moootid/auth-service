package database

import (
	"auth-service/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "microservices")
	sslMode := getEnv("DB_SSLMODE", "disable")

	// PostgreSQL connection string
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, sslMode,
	)

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// For development, you might want to use logger.Silent for production
	if getEnv("ENV", "development") == "production" {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Test connection
	if err = sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// Auto-migrate the schema
	if err := DB.AutoMigrate(&models.User{}, &models.TranscodingJob{}, &models.VideoAnalysis{}); err != nil {
		log.Fatal("Failed to auto-migrate:", err)
	}

	log.Println("Database migration completed successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
