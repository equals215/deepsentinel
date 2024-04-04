package agent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var instructionMap = map[string]func(...any) error{
	"server-address": config.ClientSetServerAddress,
}

func doConfigInstruction(instruction string, args []string) error {
	err := testIPCSocket()
	message := fmt.Sprintf("%s=%s", instruction, strings.Join(args, ","))
	log.Trace("Instruction is: ", message)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			log.Trace("Daemon not running or not acepting connections. Configuring client directly.")
			processRequest(message)
			return nil
		}
		return fmt.Errorf("failed to start IPC client: %s", err)
	}
	log.Trace("IPC Client started.")

	resp, err := sendMessageToDaemon(message)
	if err != nil {
		return fmt.Errorf("failed to send instruction to daemon: %s", err)
	}
	if resp != "ok" {
		return fmt.Errorf("unexpected response from daemon: %s", resp)
	}
	return nil
}

// ConfigCmd adds the config command to the root command
func ConfigCmd(rootCmd *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure the running agent",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configServerAddressCmd())
}

func configServerAddressCmd() *cobra.Command {
	configServerAddressCmd := &cobra.Command{
		Use:   "server-address [address]",
		Short: "Set the server address",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			config.SetLogging()
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Set server address to", args[0])
			_, err := url.Parse(args[0])
			if err != nil {
				fmt.Println("Invalid URL:", err)
				os.Exit(1)
			}
			log.Trace("URL is valid")
			err = doConfigInstruction("server-address", args)
			if err != nil {
				fmt.Println("Failed to set server address:", err)
				os.Exit(1)
			}
			log.Trace("Server address set successfully.")
		},
	}

	return configServerAddressCmd
}
