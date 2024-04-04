package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Client is the configuration for the client
var Client *ClientConfig

// ClientConfig is the configuration for the client
type ClientConfig struct {
	sync.Mutex
	ServerAddress string `json:"server_address"`
	MachineName   string `json:"machine_name"`
	LoggingLevel  string `json:"logging_level"`
	AuthToken     string `json:"auth_token"`
	MachineState  bool   `json:"machine_state"`
	ConsulState   bool   `json:"consul_state"`
	ConsulAddress string `json:"consul_address"`
	ConsulPort    string `json:"consul_port"`
	NomadState    bool   `json:"nomad_state"`
	NomadAddress  string `json:"nomad_address"`
	NomadPort     string `json:"nomad_port"`
}

// ServiceConfig is the configuration for the service
type ServiceConfig struct {
	ServiceName string `json:"service_name"`
}

// InitClient initializes the client configuration
func InitClient() {
	_initClient(true)
	SetLogging()
}

// InitClientForPanicWatcher initializes the client configuration for the panic watcher
func InitClientForPanicWatcher() {
	_initClient(false)
}

func _initClient(verbose bool) {
	Client = &ClientConfig{}
	Client.Lock()
	newClientConfig(verbose)
	Client.Unlock()
}

func RefreshClientConfig() {
	Client.Lock()
	err := Client.loadFromFile("/etc/deepsentinel/client-config.json")
	if err != nil {
		log.Errorf("failed to load client config: %s", err)
	}
	log.Trace("Client config refreshed")
	Client.Unlock()
}

func newClientConfig(verbose bool) {
	err := Client.loadFromFile("/etc/deepsentinel/client-config.json")
	if err == nil {
		return
	}

	if verbose {
		fmt.Println("Running with default configuration...")
	}

	Client.ServerAddress = "http://localhost:5000"
	Client.MachineState = true
	Client.ConsulState = false
	Client.NomadState = false
	Client.LoggingLevel = "info"

	Client.saveToFile("/etc/deepsentinel/client-config.json")
}

func PrintClientConfig(refresh ...bool) {
	refresh = append(refresh, false)
	printToLevel := func(format string, args ...interface{}) {}
	if refresh[0] == true {
		printToLevel = log.Tracef
	} else {
		printToLevel = log.Infof
		printToLevel("deepSentinel agent starting...")
	}
	printToLevel("Server address: %s\n", Client.ServerAddress)
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
