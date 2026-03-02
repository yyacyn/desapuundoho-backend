package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

var DB *sql.DB

// InitDB initializes the database connection and runs migrations
func InitDB() error {
	godotenv.Load() // Load .env, ignore error if not present

	dbHost := getEnv("DB_HOST", "")
	dbUser := getEnv("DB_USER", "")

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

	sslMode := getEnv("DB_SSLMODE", "require")
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, sslMode,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("✅ Database connected successfully!")

	if err = RunMigrations(DB, config.DBName); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// RunMigrations applies all pending migration files from the migrations/ directory
func RunMigrations(db *sql.DB, dbName string) error {
	// Source: embedded SQL files
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("error loading migration source: %w", err)
	}

	// Driver: postgres
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, dbName, driver)
	if err != nil {
		return fmt.Errorf("error creating migrator: %w", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("✅ Migrations: already up to date")
	} else {
		log.Println("✅ Migrations applied successfully!")
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// getEnv returns an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
