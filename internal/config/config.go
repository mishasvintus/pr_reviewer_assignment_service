// Package config provides application configuration loading from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host string
	Port string
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load reads configuration from environment variables.
// Returns error if required variables are not set.
func Load() (*Config, error) {
	_ = godotenv.Load()

	serverHost, err := getRequiredEnv("SERVER_HOST")
	if err != nil {
		return nil, err
	}

	serverPort, err := getRequiredEnv("SERVER_PORT")
	if err != nil {
		return nil, err
	}

	dbHost, err := getRequiredEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbPort, err := getRequiredEnv("DB_PORT")
	if err != nil {
		return nil, err
	}

	dbUser, err := getRequiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := getRequiredEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := getRequiredEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	dbSSLMode, err := getRequiredEnv("DB_SSLMODE")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: serverHost,
			Port: serverPort,
		},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
			SSLMode:  dbSSLMode,
		},
	}

	return cfg, nil
}

// DSN returns PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// getRequiredEnv reads required environment variable or returns error.
func getRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}
