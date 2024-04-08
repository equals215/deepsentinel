package config

import (
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Agent is the configuration for the agent
var Agent *AgentConfig

// AgentConfig is the configuration for the agent
type AgentConfig struct {
	sync.Mutex
	ServerAddress string `mapstructure:"server-address"`
	MachineName   string `mapstructure:"machine-name"`
	LoggingLevel  string `mapstructure:"logging-level"`
	AuthToken     string `mapstructure:"auth-token"`
	MachineState  bool   `mapstructure:"machine-state"`
}

// ServiceConfig is the configuration for the service
type ServiceConfig struct {
	ServiceName string `json:"service_name"`
}

// CraftAgentConfig parse file>env>flag for agent configuration then loads it into Agent variable
// Flags defaults set defaults for the agent configuration
func CraftAgentConfig() error {
	Agent = &AgentConfig{}

	viper.SetConfigName("agent-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/deepsentinel/")
	viper.AddConfigPath("$HOME/.deepsentinel")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	viper.SetEnvPrefix("DEEPSENTINEL")
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	Agent.Lock()
	viper.Unmarshal(&Agent)
	Agent.Unlock()

	SetLogging()

	err := viper.SafeWriteConfig()
	if err != nil && strings.Contains(err.Error(), "Already Exists") {
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

func RefreshAgentConfig() {
	Agent.Lock()
	viper.Unmarshal(&Agent)
	Agent.Unlock()

	viper.SetConfigName("agent-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/deepsentinel/")
	viper.AddConfigPath("$HOME/.deepsentinel")
	viper.AddConfigPath(".")

	log.Trace("Agent config refreshed")

	err := viper.SafeWriteConfig()
	if err != nil && strings.Contains(err.Error(), "Already Exists") {
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("failed to write config: %s", err.Error())
		}
	}
	PrintAgentConfig(true)
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
	printToLevel("Machine name: %s\n", Agent.MachineName)
}
