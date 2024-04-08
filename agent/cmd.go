package agent

import (
	"github.com/equals215/deepsentinel/config"
	"github.com/grongor/panicwatch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Those are needed because viper doesn't support same flag name accross multiple commands
// Details here: https://github.com/spf13/viper/issues/375
var (
	serverAddress string
	authToken     string
	machineName   string
	loggingLevel  string
)

// Cmd adds the agent command to the root command
func Cmd(rootCmd *cobra.Command) {
	agentCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the agent",
		PreRun: func(cmd *cobra.Command, args []string) {
			config.CraftAgentConfig()
			config.SetLogging()
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := panicwatch.Start(panicwatch.Config{
				OnPanic: func(p panicwatch.Panic) {
					reportPanic()
				},
				OnWatcherDied: func(err error) {
					log.Error("panic watcher process died")
					reportWatcherDied()
				},
			})
			if err != nil {
				log.Fatalf("failed to start panicwatch: %s", err.Error())
			}
			log.Info("Panicwatch started")
			log.Info("————————————")
			config.PrintAgentConfig()

			sock, err := startSocketServer()
			if err != nil {
				log.Fatalf("failed to start IPC socket server: %s", err.Error())
			}

			go socketIPCHandler(sock)

			work()
		},
	}
	agentCmd.Flags().StringVarP(&serverAddress, "server-address", "u", "", "Server address\nEnvironment variable: DEEPSENTINEL_SERVER_ADDRESS\n\b")
	agentCmd.Flags().StringVarP(&authToken, "auth-token", "t", "", "Auth token\nEnvironment variable: DEEPSENTINEL_AUTH_TOKEN\n\b")
	agentCmd.Flags().StringVarP(&machineName, "machine-name", "m", "", "Machine name\nEnvironment variable: DEEPSENTINEL_MACHINE_NAME\n\b")
	agentCmd.Flags().StringVarP(&loggingLevel, "logging-level", "l", "info", "Logging level\nEnvironment variable: DEEPSENTINEL_LOGGING_LEVEL\n\b")

	config.BindFlags(agentCmd.Flags())

	rootCmd.AddCommand(agentCmd)
}
