package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

// Config holds the database configuration.
type Config struct {
	Data struct {
		MasterDatabase struct {
			Driver   string `yaml:"driver" mapstructure:"driver"`
			Host     string `yaml:"host" mapstructure:"host"`
			Port     int    `yaml:"port" mapstructure:"port"`
			User     string `yaml:"user" mapstructure:"user"`
			Password string `yaml:"password" mapstructure:"password"`
			DBName   string `yaml:"dbname" mapstructure:"dbname"`
			SSLMode  string `yaml:"sslmode" mapstructure:"sslmode"`
		} `yaml:"master_database" mapstructure:"master_database"`
	} `yaml:"data" mapstructure:"data"`
}

func main() {
	// 1. Load configuration from config.yaml
	viper.SetConfigName("config.dev") // Default to dev config for local migration
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")        // For running from project root
	viper.AddConfigPath("../../configs")    // For running from cmd/migrator
	viper.AddConfigPath("../../../configs") // For go run from nested paths

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	log.Printf("Using config file: %s", viper.ConfigFileUsed())

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	dbConf := config.Data.MasterDatabase

	// 2. Expand environment variables (if any are used in config)
	host := os.ExpandEnv(dbConf.Host)
	user := os.ExpandEnv(dbConf.User)
	password := os.ExpandEnv(dbConf.Password)
	dbname := os.ExpandEnv(dbConf.DBName)
	sslmode := os.ExpandEnv(dbConf.SSLMode)
	port := dbConf.Port

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