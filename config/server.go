package config

import (
	"fmt"
	"strings"

	"github.com/equals215/deepsentinel/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Server is the configuration for the server
var Server *ServerConfig

// ServerConfig is the configuration for the server
type ServerConfig struct {
	ListeningAddress                 string `mapstructure:"address"`
	Port                             int    `mapstructure:"port"`
	AuthToken                        string `mapstructure:"auth-token"`
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
	viper.SetDefault("auth-token", "changeme")

	viper.SetConfigName("server-config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/deepsentinel/")
	viper.AddConfigPath("$HOME/.deepsentinel")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	viper.SetEnvPrefix("DEEPSENTINEL")
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if viper.Get("auth-token").(string) == "changeme" {
		fmt.Println("Generating auth token")
		viper.Set("auth-token", utils.RandStringBytesMaskImprSrcUnsafe(32))
		fmt.Printf("[WILL ONLY BE OUTPUT ONCE] Auth token: %s\n", viper.Get("auth-token"))
	}

	viper.Unmarshal(&Server)

	alertProvidersType := map[string]string{"low": viper.Get("low-alert-provider").(string), "high": viper.Get("high-alert-provider").(string)}
	alertProviders := map[string]AlertProviderConfig{"low": nil, "high": nil}
	for k, v := range alertProvidersType {
		var err error
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
			alertProviders[k] = &EmptyProvider{}
		default:
			return fmt.Errorf("'%s' is an unknown alert provider", v)
		}
	}
	Server.LowAlertProvider = alertProviders["low"]
	Server.HighAlertProvider = alertProviders["high"]

	err := viper.SafeWriteConfig()
	if strings.Contains(err.Error(), "Already Exists") {
		err := viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintServerConfig prints the server configuration
func PrintServerConfig() {
	log.Info("deepSentinel API server starting...")
	log.Infof("Serving on %s:%d", Server.ListeningAddress, Server.Port)
	log.Infof("Probe inactivity delay: %s", Server.ProbeInactivityDelay)
	log.Infof("Degraded to failed threshold: %d", Server.DegradedToFailedThreshold)
	log.Infof("Failed to alerted low threshold: %d", Server.FailedToAlertedLowThreshold)
	log.Infof("Alerted low to alerted high threshold: %d", Server.AlertedLowToAlertedHighThreshold)
}
