package config

import (
	"fmt"
	"net/url"

	"github.com/equals215/deepsentinel/utils"
	"github.com/mrz1836/go-sanitize"
	"github.com/spf13/viper"
)

// AgentSetServerAddress sets the server address in the agent config
func AgentSetServerAddress(args ...any) error {
	// Agent == nil means that CLI is doing the config change, not the running daemon
	if Agent == nil {
		CraftAgentConfig()
	}

	if len(args) == 0 {
		return fmt.Errorf("missing address")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	address := sanitize.URL(args[0].(string))
	url, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("Invalid URL: %s", err)
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		return fmt.Errorf("URL scheme is required")
	}

	Agent.Lock()
	viper.Set("server-address", address)
	Agent.Unlock()
	RefreshAgentConfig()
	return nil
}

// AgentSetAuthToken sets the auth token in the agent config
func AgentSetAuthToken(args ...any) error {
	// Agent == nil means that CLI is doing the config change, not the running daemon
	if Agent == nil {
		CraftAgentConfig()
	}

	if len(args) == 0 {
		return fmt.Errorf("missing token")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	token := utils.CleanString(args[0].(string))

	Agent.Lock()
	viper.Set("auth-token", token)
	Agent.Unlock()
	RefreshAgentConfig()
	return nil
}

// AgentSetMachineName sets the machine name in the agent config
func AgentSetMachineName(args ...any) error {
	// Agent == nil means that CLI is doing the config change, not the running daemon
	if Agent == nil {
		CraftAgentConfig()
	}

	if len(args) == 0 {
		return fmt.Errorf("missing machine name")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	name := utils.CleanString(args[0].(string))

	Agent.Lock()
	viper.Set("machine-name", name)
	Agent.Unlock()
	RefreshAgentConfig()
	return nil
}
