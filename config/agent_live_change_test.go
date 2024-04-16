package config

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func TestAgentSetServerAddress(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetLevel(log.InfoLevel)
	log.SetOutput(&logOutput)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
	})

	// Test case 1: Agent is nil
	Agent = nil
	err := AgentSetServerAddress("http://localhost:8080")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", viper.GetString("server-address"))

	// Test case 2: Missing address
	Agent = &AgentConfig{}
	err = AgentSetServerAddress()
	assert.EqualError(t, err, "missing address")

	// Test case 3: Too many arguments
	err = AgentSetServerAddress("localhost:8080", "extra-arg")
	assert.EqualError(t, err, "too many arguments")

	// Test case 4: Inalid address
	err = AgentSetServerAddress("localhost:8080")
	assert.EqualError(t, err, "URL scheme is required")

	// Test case 5: Agent is nil
	Agent = nil
	err = AgentSetServerAddress("http://localhost:8080")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", viper.GetString("server-address"))
}

func TestAgentSetAuthToken(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetLevel(log.InfoLevel)
	log.SetOutput(&logOutput)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
	})

	// Test case 1: Agent is nil
	Agent = nil
	err := AgentSetAuthToken("token123")
	assert.NoError(t, err)
	assert.Equal(t, "token123", viper.GetString("auth-token"))

	// Test case 2: Missing token
	Agent = &AgentConfig{}
	err = AgentSetAuthToken()
	assert.EqualError(t, err, "missing token")

	// Test case 3: Too many arguments
	err = AgentSetAuthToken("token123", "extra-arg")
	assert.EqualError(t, err, "too many arguments")

	// Test case 4: Valid token
	err = AgentSetAuthToken("token123")
	assert.NoError(t, err)
	assert.Equal(t, "token123", viper.GetString("auth-token"))
}

func TestAgentSetMachineName(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetLevel(log.InfoLevel)
	log.SetOutput(&logOutput)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
	})

	// Test case 1: Agent is nil
	Agent = nil
	err := AgentSetMachineName("machine1")
	assert.NoError(t, err)
	assert.Equal(t, "machine1", viper.GetString("machine-name"))

	// Test case 2: Missing machine name
	Agent = &AgentConfig{}
	err = AgentSetMachineName()
	assert.EqualError(t, err, "missing machine name")

	// Test case 3: Too many arguments
	err = AgentSetMachineName("machine1", "extra-arg")
	assert.EqualError(t, err, "too many arguments")

	// Test case 4: Valid machine name
	err = AgentSetMachineName("machine1")
	assert.NoError(t, err)
	assert.Equal(t, "machine1", viper.GetString("machine-name"))
}
