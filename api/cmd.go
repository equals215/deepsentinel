package api

import (
	"fmt"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/spf13/cobra"
)

// Cmd adds the API server command to the root command
func Cmd(rootCmd *cobra.Command) {

	apiCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the API server",
		Run: func(cmd *cobra.Command, args []string) {
			config.InitServer()
			alerting.Init()

			payloadChannel := make(chan *monitoring.Payload, 1)
			go monitoring.Handle(payloadChannel)

			addr := fmt.Sprintf("%s:%d", config.Server.ListeningAddress, config.Server.Port)
			newServer(payloadChannel).Listen(addr)
		},
	}

	rootCmd.AddCommand(apiCmd)
}
