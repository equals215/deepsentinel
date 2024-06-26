package server

import (
	"fmt"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/dashboard"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/grongor/panicwatch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd adds the API server command to the root command
func Cmd(rootCmd *cobra.Command) {
	var noAlerting bool
	var noDash bool
	payloadChannel := make(chan *monitoring.Payload)

	serverCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the API server",
		PreRun: func(cmd *cobra.Command, args []string) {
			err := config.CraftServerConfig()
			if err != nil {
				log.Fatalf("failed to craft server config: %s", err.Error())
			}
			alerting.Init(config.Server, noAlerting)
		},
		Run: func(cmd *cobra.Command, args []string) {
			var dashboardOperator *dashboard.Operator
			config.PrintServerConfig()

			if !noDash {
				dashboardOperator = dashboard.Handle()
			} else {
				log.Warn("Dashboard disabled")
			}
			go monitoring.Handle(payloadChannel, dashboardOperator)

			addr := fmt.Sprintf("%s:%d", config.Server.ListeningAddress, config.Server.Port)
			newServer(payloadChannel, dashboardOperator).Listen(addr)
			// Start panicwatch to catch panics
			err := panicwatch.Start(panicwatch.Config{
				OnPanic: func(p panicwatch.Panic) {
					alerting.ServerAlert("deepsentinel", "server", "panic")
				},
				OnWatcherDied: func(err error) {
					log.Error("panic watcher process died")
					alerting.ServerAlert("deepsentinel", "panicwatcher", "low")
				},
			})
			if err != nil {
				log.Fatalf("failed to start panicwatch: %s", err.Error())
			}
			log.Info("Panicwatch started")
		},
	}
	serverCmd.Flags().BoolVarP(&noAlerting, "no-alert", "", false, "Disable alerting")
	serverCmd.Flags().BoolVarP(&noDash, "no-dashboard", "", false, "Disable dashboard")
	serverCmd.Flags().String("address", "0.0.0.0", "Listening address\nEnvironment variable: DEEPSENTINEL_ADDRESS\n\b")
	serverCmd.Flags().String("port", "5000", "Listening port\nEnvironment variable: DEEPSENTINEL_PORT\n\b")
	serverCmd.Flags().String("probe-inactivity-delay", "2s", "Delay before considering a probe inactive\nEnvironment variable: DEEPSENTINEL_PROBE_INACTIVITY_DELAY\n\b")
	serverCmd.Flags().Int("degraded-to-failed", 10, "Number of degraded event before considering a probe or service as failed\nEnvironment variable: DEEPSENTINEL_DEGRADED_TO_FAILED\n\b")
	serverCmd.Flags().Int("failed-to-alertLow", 20, "Number of failed event before alerting low\nEnvironment variable: DEEPSENTINEL_FAILED_TO_ALERT_LOW\n\b")
	serverCmd.Flags().Int("alertLow-to-alertHigh", 30, "Number of alertLow event before alerting high\nEnvironment variable: DEEPSENTINEL_ALERT_LOW_TO_ALERT_HIGH\n\b")
	serverCmd.Flags().String("logging-level", "info", "Logging level\nEnvironment variable: DEEPSENTINEL_LOGGING_LEVEL\n\b")
	serverCmd.Flags().String("low-alert-provider", "", "Low alert provider name\nEnvironment variable: DEEPSENTINEL_LOW_ALERT_PROVIDER\n\b")
	serverCmd.Flags().String("high-alert-provider", "", "High alert provider name\nEnvironment variable: DEEPSENTINEL_HIGH_ALERT_PROVIDER\n\b")
	serverCmd.Flags().String("pagerduty.api-key", "", "PagerDuty API key\nEnvironment variable: DEEPSENTINEL_PAGERDUTY_API_KEY\n\b")
	serverCmd.Flags().String("pagerduty.integration-key", "", "PagerDuty integration key\nEnvironment variable: DEEPSENTINEL_PAGERDUTY_INTEGRATION_KEY\n\b")
	serverCmd.Flags().String("pagerduty.integration-url", "", "PagerDuty integration URL\nEnvironment variable: DEEPSENTINEL_PAGERDUTY_INTEGRATION_URL\n\b")

	config.BindFlags(serverCmd.Flags())

	rootCmd.AddCommand(serverCmd)
}
