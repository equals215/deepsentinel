package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Agent is the configuration for the agent
var Agent *AgentConfig

// AgentConfig is the configuration for the agent
type AgentConfig struct {
	sync.Mutex
	ServerAddress string `json:"server_address"`
	MachineName   string `json:"machine_name"`
	LoggingLevel  string `json:"logging_level"`
	AuthToken     string `json:"auth_token"`
	MachineState  bool   `json:"machine_state"`
}

// ServiceConfig is the configuration for the service
type ServiceConfig struct {
	ServiceName string `json:"service_name"`
}

// InitAgent initializes the agent configuration
func InitAgent() {
	_initAgent(true)
	SetLogging()
}

// InitAgentForPanicWatcher initializes the agent configuration for the panic watcher
func InitAgentForPanicWatcher() {
	_initAgent(false)
}

func _initAgent(verbose bool) {
	Agent = &AgentConfig{}
	Agent.Lock()
	newAgentConfig(verbose)
	Agent.Unlock()
}

func RefreshAgentConfig() {
	Agent.Lock()
	err := Agent.loadFromFile("/etc/deepsentinel/agent-config.json")
	if err != nil {
		log.Errorf("failed to load agent config: %s", err)
	}
	log.Trace("Agent config refreshed")
	Agent.Unlock()
}

func newAgentConfig(verbose bool) {
	err := Agent.loadFromFile("/etc/deepsentinel/agent-config.json")
	if err == nil {
		return
	}

	if verbose {
		fmt.Println("Running with default configuration...")
	}

	Agent.ServerAddress = "http://localhost:5000"
	Agent.MachineState = true
	Agent.LoggingLevel = "info"

	Agent.saveToFile("/etc/deepsentinel/agent-config.json")
}

func PrintAgentConfig(refresh ...bool) {
	refresh = append(refresh, false)
	printToLevel := func(format string, args ...interface{}) {}
	if refresh[0] == true {
		printToLevel = log.Tracef
	} else {
		printToLevel = log.Infof
		printToLevel("deepSentinel agent starting...")
	}
	printToLevel("Server address: %s\n", Agent.ServerAddress)
}

// loadFromFile loads the configuration from a JSON file
func (c *AgentConfig) loadFromFile(filePath string) error {
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

func (c *AgentConfig) saveToFile(filePath string) error {
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
