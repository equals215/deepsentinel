// Package daemonize provides a way to daemonize the deepsentinel services.
package daemonize

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/equals215/deepsentinel/config"
	"github.com/jxsl13/osfacts/distro"
)

type daemonType int

// Daemon types
const (
	Server daemonType = iota
	Agent
)

func (d daemonType) String() string {
	return [...]string{"server", "agent"}[d]
}

func (d daemonType) binaryPath() string {
	return [...]string{"/etc/deepsentinel/bin/deepsentinel-server", "/etc/deepsentinel/bin/deepsentinel-agent"}[d]
}

type daemon interface {
	isDaemonInstalled() bool
	isDaemonRunning() bool
	installDaemon() error
	uninstallDaemon() error
	updateDaemon() error
	startDaemon() error
	stopDaemon() error
}

// Daemonize installs and starts daemons
func Daemonize(component daemonType, update bool) {
	var daemon daemon
	var err error

	config.CraftAgentConfig()
	config.SetLogging()

	if daemon, err = getDaemonSystem(component); err != nil {
		fmt.Printf("Unsupported OS: %v\n", err)
		os.Exit(1)
	}
	if update {
		if !daemon.isDaemonInstalled() {
			fmt.Println("Daemon is not installed.")
			os.Exit(1)
		}
		if err := daemon.updateDaemon(); err != nil {
			fmt.Printf("Failed to update daemon: %v\n", err)
			os.Exit(1)
		}
	}
	if !daemon.isDaemonRunning() {
		if !daemon.isDaemonInstalled() {
			err := daemon.installDaemon()
			if err != nil {
				fmt.Printf("Failed to install daemon: %v\n", err)
				os.Exit(1)
			}
		}

		daemon.startDaemon()
	}
}

// UninstallDaemon uninstalls the daemon
func UninstallDaemon(component daemonType) {
	var daemon daemon
	var err error

	config.CraftAgentConfig()
	config.SetLogging()

	if daemon, err = getDaemonSystem(component); err != nil {
		fmt.Printf("Unsupported OS: %v\n", err)
		os.Exit(1)
	}
	if !daemon.isDaemonInstalled() {
		fmt.Println("Daemon is not installed.")
		os.Exit(1)
	}
	if err := daemon.uninstallDaemon(); err != nil {
		fmt.Printf("Failed to uninstall daemon: %v\n", err)
		os.Exit(1)
	}
}

func getDaemonSystem(component daemonType) (daemon, error) {
	o, err := distro.Detect()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(o.Distribution) {
	case "debian", "ubuntu", "kali":
		if _, err := os.Stat("/bin/systemctl"); err == nil {
			return &systemdDaemon{
				component: component,
			}, nil
		}
		return nil, fmt.Errorf("systemd not found")
	case "macos", "darwin":
		if _, err := os.Stat("/bin/launchctl"); err == nil {
			launchd := &launchdDaemon{
				component: component,
			}
			launchd.primeDaemon()
			return launchd, nil
		}
		return nil, fmt.Errorf("launchd not found")
	}
	return nil, fmt.Errorf("unknown or unsupported daemon system")
}

func installBinary(component daemonType) error {
	var err error

	if _, err = os.Stat(component.binaryPath()); errors.Is(err, os.ErrNotExist) {
		copyBinary(component.binaryPath())
		return nil
	}
	return err
}

func copyBinary(destination string) {
	if err := os.MkdirAll("/etc/deepsentinel/bin", 0755); err != nil {
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
