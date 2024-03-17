package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ClientConfig is the configuration for the client
type ClientConfig struct {
	ServerAddress string `json:"server_address"`
	MachineState  bool   `json:"machine_state"`
	ConsulState   bool   `json:"consul_state"`
	ConsulAddress string `json:"consul_address"`
	ConsulPort    string `json:"consul_port"`
	NomadState    bool   `json:"nomad_state"`
	NomadAddress  string `json:"nomad_address"`
	NomadPort     string `json:"nomad_port"`
}

// NewClientConfig returns a new ClientConfig
func NewClientConfig() *ClientConfig {
	config := &ClientConfig{}

	// Try to load from /etc/deepsentinel/client-config.json
	err := config.loadFromFile("/etc/deepsentinel/client-config.json")
	if err == nil {
		return config
	}

	// Try to load from .client.env
	err = config.loadFromEnvFile(".client.env")
	if err == nil {
		return config
	}

	// Try to load from environment variables
	err = config.loadFromEnv()
	if err != nil {
		// Print a message stating that it will run with defaults
		fmt.Println("Running with default configuration...")

		// Set default values for the configuration
		config.ServerAddress = "localhost"
		config.MachineState = true
		config.ConsulState = false
		config.NomadState = false
	}

	return config
}

// loadFromFile loads the configuration from a JSON file
func (c *ClientConfig) loadFromFile(filePath string) error {
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

// loadFromEnvFile loads the configuration from an environment file
func (c *ClientConfig) loadFromEnvFile(filePath string) error {
	err := godotenv.Load(filePath)
	if err != nil {
		return err
	}

	c.loadFromEnv()

	return nil
}

// loadFromEnv loads the configuration from environment variables
func (c *ClientConfig) loadFromEnv() error {
	c.ServerAddress = os.Getenv("SERVER_ADDRESS")
	machineStateStr := os.Getenv("MACHINE_STATE")
	machineState, err := strconv.ParseBool(machineStateStr)
	if err != nil {
		return err
	}
	c.MachineState = machineState
	consulStateStr := os.Getenv("CONSUL_STATE")
	consulState, err := strconv.ParseBool(consulStateStr)
	if err != nil {
		return err
	}
	c.ConsulState = consulState
	c.ConsulAddress = os.Getenv("CONSUL_ADDRESS")
	c.ConsulPort = os.Getenv("CONSUL_PORT")
	nomadStateStr := os.Getenv("NOMAD_STATE")
	nomadState, err := strconv.ParseBool(nomadStateStr)
	if err != nil {
		return err
	}
	c.NomadState = nomadState
	c.NomadAddress = os.Getenv("NOMAD_ADDRESS")
	c.NomadPort = os.Getenv("NOMAD_PORT")

	return nil
}
