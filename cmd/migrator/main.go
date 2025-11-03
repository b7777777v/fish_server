package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

// Config holds the database configuration.
type Config struct {
	Data struct {
		Database struct {
			Driver   string `yaml:"driver"`
			Host     string `yaml:"host"`
			Port     string `yaml:"port"` // Changed to string to allow env var expansion
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
			SSLMode  string `yaml:"sslmode"`
		}
	}
}

func main() {
	// 1. Load configuration from config.yaml
	viper.SetConfigName("config.dev") // Default to dev config for local migration
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	dbConf := config.Data.Database

	// 2. Expand environment variables
	host := os.ExpandEnv(dbConf.Host)
	portStr := os.ExpandEnv(dbConf.Port)
	user := os.ExpandEnv(dbConf.User)
	password := os.ExpandEnv(dbConf.Password)
	dbname := os.ExpandEnv(dbConf.DBName)
	sslmode := os.ExpandEnv(dbConf.SSLMode)

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	// 3. Construct database URL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbname,
		sslmode,
	)

	// 4. Get command-line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrator/main.go <up|down|force|version>")
	}
	command := os.Args[1]

	// 5. Initialize migrate instance
	migrationsPath := "file://storage/migrations"
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// 6. Execute command
	switch command {
	case "up":
		log.Println("Applying migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully.")
	case "down":
		log.Println("Reverting migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to revert migrations: %v", err)
		}
		log.Println("Migrations reverted successfully.")
	case "force":
		if len(os.Args) != 3 {
			log.Fatal("Usage: go run cmd/migrator/main.go force <version>")
		}
		version := 0
		fmt.Sscanf(os.Args[2], "%d", &version)
		log.Printf("Forcing migration to version %d...", version)
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force migration: %v", err)
		}
		log.Printf("Migration forced to version %d.", version)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		log.Printf("Current migration version: %d, dirty: %v", version, dirty)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}