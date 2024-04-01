package alerting

import (
	"fmt"

	"github.com/equals215/deepsentinel/alerting/providers/pagerduty"
	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
)

type AlertingConfig struct {
	lowAlertProvider  AlertProvider
	highAlertProvider AlertProvider
}

type AlertProvider interface {
	Send(category, name, status string) error
}

var Config AlertingConfig

func Init(config *config.ServerConfig) {
	var err error

	if config.LowAlertProvider != nil {
		providerConfig := config.LowAlertProvider.GetProvider()
		Config.lowAlertProvider, err = craftProvider(providerConfig)
		if err != nil {
			log.Fatal("Failed to craft low alert provider")
		}
	} else {
		Config.lowAlertProvider = nil
		log.Warn("Low alert provider is not configured")
	}

	if config.HighAlertProvider != nil {
		providerConfig := config.HighAlertProvider.GetProvider()
		Config.highAlertProvider, err = craftProvider(providerConfig)
	} else {
		Config.highAlertProvider = nil
		log.Warn("High alert provider is not configured")
	}
}

func craftProvider(provider interface{}) (AlertProvider, error) {
	if provider != nil {
		switch provider.(type) {
		case *config.PagerDutyConfig:
			log.Trace("Crafting PagerDuty provider")
			pagerDutyProvider := pagerduty.NewInstance(provider.(*config.PagerDutyConfig))
			return pagerDutyProvider, nil
		case *config.KeepHQConfig:
			log.Trace("Crafting KeepHQ provider")
			// keepHQProvider := keephqalert.NewInstance(provider.(*config.KeepHQConfig))
			return nil, nil
		default:
			return nil, fmt.Errorf("Unknown provider type")
		}
	}
	return nil, fmt.Errorf("Provider is nil")
}

func Alert(category, name, status string) {
	log.Info("Alerting ", category, " ", name, " ", status)

	if status == "low" {
		if Config.lowAlertProvider != nil {
			log.Info("Sending alert to low alert provider")
			err := Config.lowAlertProvider.Send(category, name, status)
			if err != nil {
				log.Error("Failed to send low alert: ", err)
			}
		} else {
			log.Warn("No low alert provider configured. Can't send alert.")
		}
	}

	if status == "high" {
		if Config.highAlertProvider != nil {
			log.Info("Sending alert to high alert provider")
			err := Config.highAlertProvider.Send(category, name, status)
			if err != nil {
				log.Error("Failed to send high alert: ", err)
			}
		} else {
			log.Warn("No high alert provider configured. Can't send alert.")
		}
	}
}
