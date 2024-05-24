package main

import (
	"fmt"
	"os"

	"github.com/equals215/deepsentinel/agent"
	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/daemonize"
	"github.com/equals215/deepsentinel/utils"
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
	agent.UnregisterCmd(rootCmd)
	installCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func installCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install the deepSentinel agent",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var response string

			if os.Geteuid() != 0 {
				return fmt.Errorf("you must run this command as sudo (preffered) or root")
			}

			if os.Getenv("SUDO_USER") == "" {
				fmt.Println("Running as root instead of sudo will force you to run as root in the future to configure the agent")
				fmt.Print("Do you want to continue? (y/n) ")
				fmt.Scanln(&response)
				if response != "y" {
					return fmt.Errorf("installation aborted")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var serverAddress string
			var authToken string
			var machineName string

			err := os.MkdirAll("/etc/deepsentinel/", 0644)
			if err != nil {
				return err
			}

			fmt.Print("Input the server address (ex.: http://localhost:8080): ")
			fmt.Scanln(&serverAddress)
			err = config.AgentSetServerAddress(utils.CleanString(serverAddress))
			if err != nil {
				return err
			}

			fmt.Print("Input the auth token: ")
			fmt.Scanln(&authToken)
			err = config.AgentSetAuthToken(utils.CleanString(authToken))
			if err != nil {
				return err
			}

			fmt.Print("Input the machine name: ")
			fmt.Scanln(&machineName)
			err = config.AgentSetMachineName(utils.CleanString(machineName))
			if err != nil {
				return err
			}

			daemonize.Daemonize(daemonize.Agent, false)
			return nil
		},
	})
}
