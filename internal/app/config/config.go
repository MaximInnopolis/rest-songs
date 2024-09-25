package config

import (
	"fmt"
	"os"
)

var defaultHttpPort = ":8080"

// Config struct holds configuration values for database url and http port
type Config struct {
	DbUrl    string
	HttpPort string
}

// New creates new Config instance by reading environment variables
// It checks if required DATABASE_URL is set; if not, it returns error
// If HTTP_PORT is not set, it defaults to ":8080".
func New() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL не задан")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = defaultHttpPort
	}

	return &Config{
		DbUrl:    dbURL,
		HttpPort: httpPort,
	}, nil
}
