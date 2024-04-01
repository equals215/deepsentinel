package pagerdutyalert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/equals215/deepsentinel/config"
)

type PagerDutyInstance struct {
	config *config.PagerDutyConfig
}

type AlertPayload struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
}

// NewPagerDutyInstance creates a new PagerDuty instance
func NewPagerDutyInstance(config *config.PagerDutyConfig) *PagerDutyInstance {
	return &PagerDutyInstance{
		config: config,
	}
}

// SendAlert sends an alert to PagerDuty
func (instance PagerDutyInstance) SendAlert(componentType, component, severity string) error {
	if componentType == "machine" {
		summary := fmt.Sprintf("Machine %s is %s", component, severity)
		return _sendPagerDutyAlert(instance, summary, severity)
	} else if componentType == "service" {
		summary := fmt.Sprintf("Service %s is %s", component, severity)
		return _sendPagerDutyAlert(instance, summary, severity)
	}
	summary := fmt.Sprintf("Unknown component %s is %s", component, severity)
	return _sendPagerDutyAlert(instance, summary, severity)
}

func _sendPagerDutyAlert(instance PagerDutyInstance, summary string, severity string) error {
	payload := AlertPayload{
		Summary:  summary,
		Severity: severity,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.pagerduty.com/alerts", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", instance.config.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send PagerDuty alert: %s", resp.Status)
	}

	return nil
}
