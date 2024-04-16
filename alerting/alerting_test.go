package alerting

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func TestServerAlert(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
	})

	// Test case 1: severity is "low" and lowAlertProvider is configured
	Config.lowAlertProvider = &MockAlertProvider{}
	ServerAlert("category", "component", "low")
	assert.Equal(t, "Sending alert to low alert provider: MockProvider", logOutput.String())
	logOutput.Reset()

	// Test case 2: severity is "low" but lowAlertProvider is not configured
	Config.lowAlertProvider = nil
	ServerAlert("category", "component", "low")
	assert.Equal(t, "No low alert provider configured. Can't send alert.", logOutput.String())
	logOutput.Reset()

	// Test case 3: severity is "high" and highAlertProvider is configured
	Config.highAlertProvider = &MockAlertProvider{}
	ServerAlert("category", "component", "high")
	assert.Equal(t, "Sending alert to high alert provider: MockProvider", logOutput.String())
	logOutput.Reset()

	// Test case 4: severity is "high" but highAlertProvider is not configured
	Config.highAlertProvider = nil
	ServerAlert("category", "component", "high")
	assert.Equal(t, "No high alert provider configured. Can't send alert.", logOutput.String())
	logOutput.Reset()

	// Test case 5: severity is "panic" and highAlertProvider is configured
	Config.highAlertProvider = &MockAlertProvider{}
	ServerAlert("category", "component", "panic")
	assert.Equal(t, "Sending alert to high alert provider: MockProvider", logOutput.String())
	logOutput.Reset()

	// Test case 6: severity is "panic" but highAlertProvider is not configured
	Config.highAlertProvider = nil
	ServerAlert("category", "component", "panic")
	assert.Equal(t, "No high alert provider configured. Can't send alert.", logOutput.String())
	logOutput.Reset()
}

// MockAlertProvider is a mock implementation of the AlertProvider interface
type MockAlertProvider struct{}

func (m *MockAlertProvider) Name() string {
	return "MockProvider"
}

func (m *MockAlertProvider) Send(category, component, severity string) error {
	return nil
}
