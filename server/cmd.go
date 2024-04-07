package server

import (
	"fmt"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config/v1"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/grongor/panicwatch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd adds the API server command to the root command
func Cmd(rootCmd *cobra.Command) {
	var noAlerting bool

	serverCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the API server",
		PreRun: func(cmd *cobra.Command, args []string) {
			config.InitServerForPanicWatcher()
			alerting.InitForPanicWatcher(config.Server, noAlerting)

			//Start panicwatch to catch panics
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
		Run: func(cmd *cobra.Command, args []string) {
			config.InitServer()
			alerting.Init(config.Server, noAlerting)

			log.Infof("————————————")

			config.PrintServerConfig()
			payloadChannel := make(chan *monitoring.Payload, 1)
			go monitoring.Handle(payloadChannel)

			addr := fmt.Sprintf("%s:%d", config.Server.ListeningAddress, config.Server.Port)
			newServer(payloadChannel).Listen(addr)
		},
	}
	serverCmd.Flags().BoolVarP(&noAlerting, "no-alert", "", false, "Disable alerting")

	rootCmd.AddCommand(serverCmd)
}
