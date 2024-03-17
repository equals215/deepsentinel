package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/equals215/deepsentinel/utils"
)

// Server is the configuration for the server
var Server *ServerConfig

// ServerConfig is the configuration for the server
type ServerConfig struct {
	ListeningAddress string `json:"listening_address"`
	Port             int    `json:"port"`
	AuthToken        string `json:"auth_token"`
}

// InitServer initializes the server configuration
func InitServer() {
	Server = newServerConfig()
}

func newServerConfig() *ServerConfig {
	config := &ServerConfig{}

	err := config.loadFromFile("/etc/deepsentinel/server-config.json")
	if err == nil {
		return config
	}

	fmt.Println("Running with default configuration...")

	config.ListeningAddress = "localhost"
	config.Port = 5000
	config.AuthToken = utils.RandStringBytesMaskImprSrcUnsafe(32)

	err = config.saveToFile("/etc/deepsentinel/server-config.json")
	if err != nil {
		fmt.Println("Error saving configuration file:", err)
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

func (c *ServerConfig) saveToFile(filePath string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll("/etc/deepsentinel", 0755); err != nil {
			return err
		}
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
