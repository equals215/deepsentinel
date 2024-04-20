package pagerduty

import (
	"fmt"
	"testing"

	"github.com/equals215/deepsentinel/config"
	"github.com/stretchr/testify/assert"
)

func Test_sendPagerDutyAlert(t *testing.T) {
	instance := PagerDutyInstance{
		config: &config.PagerDutyConfig{
			IntegrationKey: "test-integration-key",
		},
		client: nil, // Replace with your PagerDuty client implementation
	}

	summary := "Test Summary"
	component := "Test Component"
	severity := "low"

	err := _sendPagerDutyAlert(instance, summary, component, severity)
	assert.NoError(t, err, "Expected no error for low severity")

	severity = "high"
	err = _sendPagerDutyAlert(instance, summary, component, severity)
	assert.NoError(t, err, "Expected no error for high severity")

	severity = "unknown"
	err = _sendPagerDutyAlert(instance, summary, component, severity)
	assert.Error(t, err, fmt.Sprintf("Expected error for unknown severity %s", severity))
}
