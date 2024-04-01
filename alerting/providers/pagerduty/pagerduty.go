package pagerduty

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	pagerdutysdk "github.com/PagerDuty/go-pagerduty"
	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/utils"
	log "github.com/sirupsen/logrus"
)

type PagerDutyInstance struct {
	config *config.PagerDutyConfig
	client *pagerdutysdk.Client
}

type AlertPayload struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
}

// NewInstance creates a new PagerDuty instance
func NewInstance(config *config.PagerDutyConfig) PagerDutyInstance {
	log.Info("Warming PagerDuty instance")

	pdCLient := pagerdutysdk.NewClient(config.APIKey)
	pdInstance := PagerDutyInstance{
		config: config,
		client: pdCLient,
	}
	if pdInstance.client != nil {
		pdInstance.client.SetDebugFlag(pagerdutysdk.DebugCaptureLastResponse)
	}
	if err := pdInstance.ping(); err != nil {
		log.Fatalf("PagerDuty ping failed: %v", err)
		return PagerDutyInstance{}
	}
	return pdInstance
}

// Send sends an alert to PagerDuty
func (instance PagerDutyInstance) Send(category, component, severity string) error {
	if category == "machine" {
		summary := fmt.Sprintf("Deepsentinel - Machine %s alert level is %s", component, severity)
		return _sendPagerDutyAlert(instance, summary, component, severity)
	} else if category == "service" {
		summary := fmt.Sprintf("Deepsentinel - Service %s alert level is %s", component, severity)
		return _sendPagerDutyAlert(instance, summary, component, severity)
	}
	summary := fmt.Sprintf("Unknown component %s is %s", component, severity)
	return _sendPagerDutyAlert(instance, summary, component, severity)
}

func _sendPagerDutyAlert(instance PagerDutyInstance, summary, component, severity string) error {
	ctx := context.Background()

	if severity == "low" {
		severity = "warning"
	} else if severity == "high" {
		severity = "critical"
	} else {
		return fmt.Errorf("unknown severity %s", severity)
	}

	// Construct the event details
	event := &pagerdutysdk.V2Event{
		RoutingKey: instance.config.IntegrationKey,
		Action:     "trigger",
		DedupKey:   utils.RandStringBytesMaskImprSrcUnsafe(12),
		Payload: &pagerdutysdk.V2Payload{
			Summary:   summary,
			Source:    component,
			Severity:  severity,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Send the event
	response, err := instance.client.ManageEventWithContext(ctx, event)
	if err != nil {
		var httpRespBody []byte
		httpResp, _ := instance.client.LastAPIResponse()
		if httpResp != nil && httpResp.Body != nil {
			httpRespBody, _ = io.ReadAll(httpResp.Body)
		}
		var outErr error
		if response != nil {
			outErr = fmt.Errorf("%s problem creating PagerDuty event caused by %s: %s", response.Status, response.Message, strings.Join(response.Errors, ", "))
		} else {
			outErr = fmt.Errorf("unexpected response of %s from creating event in PagerDuty: %v", string(httpRespBody), err)
		}
		return outErr
	}

	log.Infof("Event sent to PagerDuty successfully: %s", response.Status)
	return nil
}

func (instance *PagerDutyInstance) ping() error {
	if instance == nil || instance.client == nil {
		return errors.New("pagerduty: nil")
	}

	ctx := context.Background()
	resp, err := instance.client.ListAbilitiesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("pagerduty list abilities: %v", err)
	}
	if len(resp.Abilities) <= 0 {
		return fmt.Errorf("pagerduty: missing abilities")
	}

	return nil
}
