package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Try to load .env file, but don't fail if it doesn't exist
	godotenv.Load() // Ignore any errors

	// Get config from environment variables
	dbHost := getEnv("DB_HOST", "")
	dbUser := getEnv("DB_USER", "")

	// Skip database connection if no credentials provided
	if dbHost == "" || dbUser == "" {
		log.Println("⚠️  No database credentials found, skipping database connection")
		return nil
	}

	config := DBConfig{
		Host:     dbHost,
		Port:     getEnv("DB_PORT", "5432"),
		User:     dbUser,
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "postgres"),
	}

	// Build connection string
	// Use sslmode=require for cloud databases like NeonDB
	// Use sslmode=disable for local/cPanel databases
	sslMode := getEnv("DB_SSLMODE", "require")
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, sslMode,
	)

	// Open database connection
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("✅ Database connected successfully!")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
