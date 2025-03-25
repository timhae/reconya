package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type DatabaseType string

const (
	MongoDB DatabaseType = "mongodb"
	SQLite  DatabaseType = "sqlite"
)

type Config struct {
	JwtKey       []byte
	NetworkCIDR  string
	DatabaseType DatabaseType
	// MongoDB config
	MongoURI     string
	// SQLite config
	SQLitePath   string
	// Common configs
	Username     string
	Password     string
	DatabaseName string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	networkCIDR := os.Getenv("NETWORK_RANGE")
	if networkCIDR == "" {
		return nil, fmt.Errorf("NETWORK_RANGE is not set in .env file")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		return nil, fmt.Errorf("DATABASE_NAME is not set in .env file")
	}

	username := os.Getenv("LOGIN_USERNAME")
	password := os.Getenv("LOGIN_PASSWORD")
	if username == "" || password == "" {
		return nil, fmt.Errorf("LOGIN_USERNAME or LOGIN_PASSWORD is not set in .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY is not set in .env file")
	}

	// Determine database type
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		dbType = string(SQLite) // Default to SQLite
	}

	config := &Config{
		JwtKey:       []byte(jwtSecret),
		NetworkCIDR:  networkCIDR,
		DatabaseType: DatabaseType(dbType),
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
	}

	// Configure based on database type
	if config.DatabaseType == MongoDB {
		mongoURI := os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			return nil, fmt.Errorf("MONGODB_URI is not set in .env file")
		}
		config.MongoURI = mongoURI
	} else if config.DatabaseType == SQLite {
		sqlitePath := os.Getenv("SQLITE_PATH")
		if sqlitePath == "" {
			// Default to a data directory in the current directory
			sqlitePath = filepath.Join("data", fmt.Sprintf("%s.db", databaseName))
		}
		config.SQLitePath = sqlitePath
	} else {
		return nil, fmt.Errorf("unsupported DATABASE_TYPE: %s", dbType)
	}

	return config, nil
}
