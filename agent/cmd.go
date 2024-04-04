package agent

import (
	"github.com/equals215/deepsentinel/config"
	"github.com/grongor/panicwatch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd adds the agent command to the root command
func Cmd(rootCmd *cobra.Command) {
	agentCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the agent",
		Run: func(cmd *cobra.Command, args []string) {
			config.InitClient()
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
			config.PrintClientConfig()

			sock, err := startSocketServer()
			if err != nil {
				log.Fatalf("failed to start IPC socket server: %s", err.Error())
			}
			defer sock.Close()

			go socketIPCHandler(sock)

			work()
		},
	}

	rootCmd.AddCommand(agentCmd)
}
