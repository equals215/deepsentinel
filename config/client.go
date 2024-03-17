package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Client is the configuration for the client
var Client *ClientConfig

// ClientConfig is the configuration for the client
type ClientConfig struct {
	ServerAddress string `json:"server_address"`
	AuthToken     string `json:"auth_token"`
	MachineState  bool   `json:"machine_state"`
	ConsulState   bool   `json:"consul_state"`
	ConsulAddress string `json:"consul_address"`
	ConsulPort    string `json:"consul_port"`
	NomadState    bool   `json:"nomad_state"`
	NomadAddress  string `json:"nomad_address"`
	NomadPort     string `json:"nomad_port"`
}

// InitClient initializes the client configuration
func InitClient() {
	Client = newClientConfig()
}

func newClientConfig() *ClientConfig {
	config := &ClientConfig{}

	err := config.loadFromFile("/etc/deepsentinel/client-config.json")
	if err == nil {
		return config
	}

	fmt.Println("Running with default configuration...")

	config.ServerAddress = "localhost:5000"
	config.MachineState = true
	config.ConsulState = false
	config.NomadState = false

	config.saveToFile("/etc/deepsentinel/client-config.json")

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

func (c *ClientConfig) saveToFile(filePath string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
