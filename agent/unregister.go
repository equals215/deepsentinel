package agent

import (
	"fmt"

	"github.com/spf13/cobra"
)

// UnregisterCmd provides unregister cli command
func UnregisterCmd(rootCmd *cobra.Command) {
	unregisterCmd := &cobra.Command{
		Use:   "unregister",
		Short: "Unregister and stops the running agent",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Unregistering agent")
			doConfigInstruction("unregister", args)
		},
	}

	rootCmd.AddCommand(unregisterCmd)
}
