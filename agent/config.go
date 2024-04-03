package agent

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func ConfigCmd(rootCmd *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure the running agent",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Configure instance")
		},
	}

	rootCmd.AddCommand(configCmd)
}
