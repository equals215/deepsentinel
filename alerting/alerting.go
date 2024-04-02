package alerting

import (
	"fmt"
	"io"

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
	Name() string
}

var Config AlertingConfig

func Init(config *config.ServerConfig, noAlerting bool) {
	_init(config, true, noAlerting)
}

func InitForPanicWatcher(serverConfig *config.ServerConfig, noAlerting bool) {
	log.SetOutput(io.Discard)

	_init(serverConfig, false, noAlerting)

	config.SetLogging()
}

func _init(serverConfig *config.ServerConfig, verbose bool, noAlerting bool) {
	var err error

	if serverConfig.LowAlertProvider != nil {
		providerConfig := serverConfig.LowAlertProvider.GetProvider()
		Config.lowAlertProvider, err = craftProvider(providerConfig)
		if err != nil {
			log.Fatal("Failed to craft low alert provider")
		}
	} else {
		Config.lowAlertProvider = nil
		log.Warn("Low alert provider is not configured")
	}

	if serverConfig.HighAlertProvider != nil {
		providerConfig := serverConfig.HighAlertProvider.GetProvider()
		Config.highAlertProvider, err = craftProvider(providerConfig)
	} else {
		Config.highAlertProvider = nil
		log.Warn("High alert provider is not configured")
	}

	if noAlerting {
		Config.lowAlertProvider = nil
		Config.highAlertProvider = nil
		log.Warn("Alerting is disabled due to -no-alert flag")
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

func ServerAlert(category, component, severity string) {
	log.Tracef("Alerting %s %s %s", category, component, severity)

	if severity == "low" {
		if Config.lowAlertProvider != nil {
			log.Infof("Sending alert to low alert provider: %s", Config.lowAlertProvider.Name())
			err := Config.lowAlertProvider.Send(category, component, severity)
			if err != nil {
				log.Error("Failed to send low alert: ", err)
			}
		} else {
			log.Warn("No low alert provider configured. Can't send alert.")
		}
	}

	if severity == "high" {
		if Config.highAlertProvider != nil {
			log.Infof("Sending alert to high alert provider: %s", Config.highAlertProvider.Name())
			err := Config.highAlertProvider.Send(category, component, severity)
			if err != nil {
				log.Error("Failed to send high alert: ", err)
			}
		} else {
			log.Warn("No high alert provider configured. Can't send alert.")
		}
	}

	if severity == "panic" {
		if Config.highAlertProvider != nil {
			log.Infof("Sending alert to high alert provider: %s", Config.highAlertProvider.Name())
			err := Config.highAlertProvider.Send(category, component, severity)
			if err != nil {
				log.Error("Failed to send panic alert: ", err)
			}
		} else {
			log.Warn("No high alert provider configured. Can't send alert.")
		}
	}
}
