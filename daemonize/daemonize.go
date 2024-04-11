// Package daemonize provides a way to daemonize the deepsentinel services.
package daemonize

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jxsl13/osfacts/distro"
)

type daemonType int

// Daemon types
const (
	Server daemonType = iota
	Agent
)

var (
	serverBinaryPath = "/etc/deepsentinel/server"
	agentBinaryPath  = "/etc/deepsentinel/client"
)

type daemon interface {
	isDaemonInstalled(daemonType) bool
	isDaemonRunning(daemonType) bool
	installDaemon(daemonType) error
	uninstallDaemon(daemonType) error
	updateDaemon(daemonType) error
	startDaemon(daemonType) error
	stopDaemon(daemonType) error
}

// Daemonize installs and starts daemons
func Daemonize(component daemonType, update bool) {
	var daemon daemon
	var err error

	if daemon, err = getDaemonSystem(); err != nil {
		fmt.Println("Unsupported OS.")
		os.Exit(1)
	}
	if update {
		if !daemon.isDaemonInstalled(component) {
			fmt.Println("Daemon is not installed.")
			os.Exit(1)
		}
		if err := daemon.updateDaemon(component); err != nil {
			fmt.Printf("Failed to update daemon: %v\n", err)
			os.Exit(1)
		}
	}
	if !daemon.isDaemonRunning(component) {
		if component == Agent {
			if _, err := os.Stat(agentBinaryPath); errors.Is(err, os.ErrNotExist) {
				copyBinary(agentBinaryPath)
			}
		} else if component == Server {
			if _, err := os.Stat(serverBinaryPath); errors.Is(err, os.ErrNotExist) {
				copyBinary(serverBinaryPath)
			}
		}
		if !daemon.isDaemonInstalled(component) {
			err := daemon.installDaemon(component)
			if err != nil {
				fmt.Printf("Failed to install daemon: %v\n", err)
				os.Exit(1)
			}
		}

		daemon.startDaemon(component)
	}
}

// UninstallDaemon uninstalls the daemon
func UninstallDaemon(component daemonType) {
	var daemon daemon
	var err error

	if daemon, err = getDaemonSystem(); err != nil {
		fmt.Println("Unsupported OS.")
		os.Exit(1)
	}
	if !daemon.isDaemonInstalled(component) {
		fmt.Println("Daemon is not installed.")
		os.Exit(1)
	}
	if err := daemon.uninstallDaemon(component); err != nil {
		fmt.Printf("Failed to uninstall daemon: %v\n", err)
		os.Exit(1)
	}
}

func getDaemonSystem() (daemon, error) {
	o, err := distro.Detect()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(o.Distribution) {
	case "debian", "ubuntu", "kali":
		if _, err := os.Stat("/bin/systemctl"); err == nil {
			return &systemdDaemon{}, nil
		}
		return nil, fmt.Errorf("systemd not found")
	}
	return nil, fmt.Errorf("unknown or unsupported daemon system")
}

func copyBinary(destination string) {
	if err := os.MkdirAll("/etc/deepsentinel", 0755); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get binary path: %v\n", err)
		os.Exit(1)
	}

	if err := os.Rename(binaryPath, destination); err != nil {
		fmt.Printf("Failed to copy binary: %v\n", err)
		os.Exit(1)
	}
}
