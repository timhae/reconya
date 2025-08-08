package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type DatabaseType string

const (
	SQLite DatabaseType = "sqlite"
)

type Config struct {
	JwtKey       []byte
	Port         string
	DatabaseType DatabaseType
	// SQLite config
	SQLitePath string
	// Common configs
	Username     string
	Password     string
	DatabaseName string
}

func LoadConfig() (*Config, error) {
	// Try to load .env file but don't fail if it doesn't exist
	// This allows using environment variables directly in Docker
	_ = godotenv.Load()

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		return nil, fmt.Errorf("DATABASE_NAME environment variable is not set")
	}

	username := os.Getenv("LOGIN_USERNAME")
	password := os.Getenv("LOGIN_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("LOGIN_USERNAME or LOGIN_PASSWORD environment variables are not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY environment variable is not set")
	}

	// Get port (default to 3008)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3008"
	}

	// Set database type to SQLite
	dbType := string(SQLite)

	config := &Config{
		JwtKey:       []byte(jwtSecret),
		Port:         port,
		DatabaseType: DatabaseType(dbType),
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
	}

	// Configure SQLite database
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		// Default to a data directory in the current directory
		sqlitePath = filepath.Join("data", fmt.Sprintf("%s.db", databaseName))
	}
	config.SQLitePath = sqlitePath

	return config, nil
}
