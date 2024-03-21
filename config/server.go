package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/equals215/deepsentinel/utils"
	log "github.com/sirupsen/logrus"
)

// Server is the configuration for the server
var Server *ServerConfig

// ServerConfig is the configuration for the server
type ServerConfig struct {
	ListeningAddress                 string `json:"listening_address"`
	Port                             int    `json:"port"`
	AuthToken                        string `json:"auth_token"`
	ProbeInactivityDelaySeconds      int    `json:"probe_inactivity_delay_seconds"`
	DegradedToFailedThreshold        int    `json:"degraded_to_failed_threshold"`
	FailedToAlertedLowThreshold      int    `json:"failed_to_alerted_low_threshold"`
	AlertedLowToAlertedHighThreshold int    `json:"alerted_low_to_alerted_high_threshold"`
}

// InitServer initializes the server configuration
func InitServer() {
	Server = newServerConfig()
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Info("deepSentinel API server starting...")
	log.Infof("Serving on %s:%d", Server.ListeningAddress, Server.Port)
	log.Infof("Probe inactivity delay: %d seconds", Server.ProbeInactivityDelaySeconds)
	log.Infof("Degraded to failed threshold: %d", Server.DegradedToFailedThreshold)
	log.Infof("Failed to alerted low threshold: %d", Server.FailedToAlertedLowThreshold)
	log.Infof("Alerted low to alerted high threshold: %d", Server.AlertedLowToAlertedHighThreshold)
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
	config.ProbeInactivityDelaySeconds = 5
	config.DegradedToFailedThreshold = 10
	config.FailedToAlertedLowThreshold = 10
	config.AlertedLowToAlertedHighThreshold = 10

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
