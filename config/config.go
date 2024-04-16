package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindFlags binds the flags to the viper configuration
// This is needed because viper doesn't support same flag name accross multiple commands
// Details here: https://github.com/spf13/viper/issues/375#issuecomment-794668149
func BindFlags(flagSet *pflag.FlagSet) {
	flagSet.VisitAll(func(flag *pflag.Flag) {
		viper.BindPFlag(flag.Name, flag)
	})
}

// SetLogging sets the logging level and format
func SetLogging() {
	var loggingLevel string

	log.SetOutput(os.Stdout)

	if Server != nil {
		loggingLevel = Server.LoggingLevel
	} else if Agent != nil {
		loggingLevel = Agent.LoggingLevel
	}
	if loggingLevel == "" {
		loggingLevel = "info"
	}

	logLevel, err := log.ParseLevel(loggingLevel)
	if err != nil {
		log.Fatalf("couldn't set logging level: %s", err)
	}
	log.SetLevel(logLevel)

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}
