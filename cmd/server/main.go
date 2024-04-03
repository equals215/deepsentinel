package main

import (
	"fmt"
	"os"

	"github.com/equals215/deepsentinel/daemonize"
	"github.com/equals215/deepsentinel/server"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deepsentinel-server",
		Short: "deepSentinel server CLI",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	daemonize.Cmd(rootCmd, daemonize.Server)
	server.Cmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
