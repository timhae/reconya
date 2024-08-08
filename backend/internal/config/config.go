package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JwtKey       []byte
	NetworkCIDR  string
	MongoURI     string
	Username     string
	Password     string
	DatabaseName string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGODB_URI is not set in .env file")
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

	return &Config{
		JwtKey:       []byte("my_secret_key"),
		NetworkCIDR:  networkCIDR,
		MongoURI:     mongoURI,
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
	}, nil
}
