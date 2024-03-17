package api

import (
	"fmt"

	"github.com/equals215/deepsentinel/config"
	"github.com/spf13/cobra"
)

// Cmd adds the API server command to the root command
func Cmd(rootCmd *cobra.Command) {
	config.InitServer()

	apiCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the API server",
		Run: func(cmd *cobra.Command, args []string) {
			addr := fmt.Sprintf("%s:%d", config.Server.ListeningAddress, config.Server.Port)
			newServer().Listen(addr)
		},
	}

	rootCmd.AddCommand(apiCmd)
}
