package daemonize

import (
	"os"

	"github.com/spf13/cobra"
)

// Cmd adds the daemonize commands to the root command
func Cmd(rootCmd *cobra.Command, component daemonType) {
	var componentStr string
	if component == Server {
		componentStr = "server"
	} else if component == Agent {
		componentStr = "agent"
	}

	daemonizeCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Daemonize the deepSentinel " + componentStr,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd.AddCommand(daemonizeCmd)

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install the deepSentinel " + componentStr + " as a daemon",
		Run: func(cmd *cobra.Command, args []string) {
			Daemonize(component, false)
		},
	}
	daemonizeCmd.AddCommand(installCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the deepSentinel " + componentStr + " daemon",
		Run: func(cmd *cobra.Command, args []string) {
			Daemonize(component, true)
		},
	}
	daemonizeCmd.AddCommand(updateCmd)

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the deepSentinel " + componentStr + " daemon",
		Run: func(cmd *cobra.Command, args []string) {
			UninstallDaemon(component)
		},
	}
	daemonizeCmd.AddCommand(uninstallCmd)
}
