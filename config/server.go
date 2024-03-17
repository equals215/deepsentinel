package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ServerConfig is the configuration for the server
type ServerConfig struct {
	// Port is the port the server listens on
	Port int `json:"port"`
}

// NewServerConfig returns a new ServerConfig
func NewServerConfig() *ServerConfig {
	config := &ServerConfig{}

	// Try to load from /etc/deepsentinel/client-config.json
	err := config.loadFromFile("/etc/deepsentinel/server-config.json")
	if err == nil {
		return config
	}

	// Try to load from .client.env
	err = config.loadFromEnvFile(".server.env")
	if err == nil {
		return config
	}

	// Try to load from environment variables
	err = config.loadFromEnv()
	if err != nil {
		// Print a message stating that it will run with defaults
		fmt.Println("Running with default configuration...")

		// Set default values for the configuration
		config.Port = 5000
	}

	return config
}

func (c *ServerConfig) loadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServerConfig) loadFromEnvFile(filePath string) error {
	err := godotenv.Load(filePath)
	if err != nil {
		return err
	}

	c.loadFromEnv()

	return nil
}

func (c *ServerConfig) loadFromEnv() error {
	var err error

	c.Port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return err
	}
	return nil
}
