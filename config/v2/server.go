package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Server is the configuration for the server
var Server *ServerConfig

// ServerConfig is the configuration for the server
type ServerConfig struct {
	ListeningAddress                 string `mapstructure:"address"`
	Port                             int    `mapstructure:"port"`
	AuthToken                        string
	ProbeInactivityDelay             string `mapstructure:"probe-inactivity-delay"`
	DegradedToFailedThreshold        int    `mapstructure:"degraded-to-failed"`
	FailedToAlertedLowThreshold      int    `mapstructure:"failed-to-alertLow"`
	AlertedLowToAlertedHighThreshold int    `mapstructure:"alertLow-to-alertHigh"`
	LoggingLevel                     string `mapstructure:"logging-level"`
	LowAlertProvider                 AlertProviderConfig
	HighAlertProvider                AlertProviderConfig
}

// ServerBindFlags binds the server configuration flags to viper flagset
func ServerBindFlags(flagSet *pflag.FlagSet) {
	viper.BindPFlags(flagSet)
}

func craftAlertProviderConfig(a AlertProviderType) (AlertProviderConfig, error) {
	switch a.String() {
	case "pagerduty":
		return &PagerDutyConfig{
			APIKey:         viper.GetString("pagerduty.api-key"),
			IntegrationKey: viper.GetString("pagerduty.integration-key"),
			IntegrationURL: viper.GetString("pagerduty.integration-url"),
		}, nil
	case "keephq":
		return nil, fmt.Errorf("keephq provider is not implemented")
	default:
		return nil, fmt.Errorf("unknown provider")
	}
}

// CraftServerConfig parse file>env>flag for server configuration then loads it into Server variable
// Flags defaults set defaults for the server configuration
func CraftServerConfig() error {
	viper.SetConfigName("server-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/deepsentinel/")
	viper.AddConfigPath("$HOME/.deepsentinel")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	viper.SetEnvPrefix("DEEPSENTINEL")
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	viper.Unmarshal(&Server)

	alertProvidersType := map[string]string{"low": viper.Get("low-alert-provider").(string), "high": viper.Get("high-alert-provider").(string)}
	alertProviders := map[string]AlertProviderConfig{"low": nil, "high": nil}
	for k, v := range alertProvidersType {
		switch v {
		case "pagerduty":
			alertProviders[k], err = craftAlertProviderConfig(pagerDuty)
			if err != nil {
				return err
			}
		case "keephq":
			alertProviders[k], err = craftAlertProviderConfig(keepHQ)
			if err != nil {
				return err
			}
		case "":
			alertProviders[k] = nil
		default:
			return fmt.Errorf("'%s' is an unknown alert provider", v)
		}
	}
	Server.LowAlertProvider = alertProviders["low"]
	Server.HighAlertProvider = alertProviders["high"]

	return nil
}
