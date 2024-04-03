package main

import (
	"fmt"
	"os"

	"github.com/equals215/deepsentinel/agent"
	"github.com/equals215/deepsentinel/daemonize"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deepsentinel",
		Short: "deepSentinel agent CLI",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	daemonize.Cmd(rootCmd, daemonize.Agent)
	agent.Cmd(rootCmd)
	agent.ConfigCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
