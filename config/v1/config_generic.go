package config

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// SetLogging sets the logging level and format
func SetLogging() {
	var loggingLevel string

	log.SetOutput(os.Stdout)

	if Server != nil {
		loggingLevel = Server.LoggingLevel
	} else if Agent != nil {
		loggingLevel = Agent.LoggingLevel
	} else if os.Getenv("LOG_LEVEL") != "" {
		loggingLevel = os.Getenv("LOG_LEVEL")
	} else {
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
