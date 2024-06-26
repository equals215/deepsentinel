package daemonize

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Cmd adds the daemonize commands to the root command
func Cmd(rootCmd *cobra.Command, component daemonType) {
	daemonizeCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Daemonize the deepSentinel " + component.String() + ". Must be run as root.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.AddCommand(daemonizeCmd)

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install the deepSentinel " + component.String() + " as a daemon",
		Run: func(cmd *cobra.Command, args []string) {
			Daemonize(component, false)
			fmt.Println("Daemon installed.")
			fmt.Println("If you didn't add deepSentinel to the system PATH, you follow the following steps to add the binaries to yours.")
			fmt.Println("For Bash:")
			fmt.Println("Add `export PATH=$PATH:/etc/deepsentinel/bin` to your ~/.bashrc file and run `source ~/.bashrc`.")
			fmt.Println("For Zsh:")
			fmt.Println("Add `export PATH=$PATH:/etc/deepsentinel/bin` to your ~/.zshrc file and run `source ~/.zshrc`.")
		},
	}
	daemonizeCmd.AddCommand(installCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the deepSentinel " + component.String() + " daemon",
		Run: func(cmd *cobra.Command, args []string) {
			Daemonize(component, true)
			fmt.Println("Daemon updated.")
		},
	}
	daemonizeCmd.AddCommand(updateCmd)

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the deepSentinel " + component.String() + " daemon",
		Run: func(cmd *cobra.Command, args []string) {
			UninstallDaemon(component)
			fmt.Println("Daemon uninstalled.")
		},
	}
	daemonizeCmd.AddCommand(uninstallCmd)
}
