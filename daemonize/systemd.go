package daemonize

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/equals215/deepsentinel/agent"
)

func (d systemdDaemon) serviceFileName() string {
	return [...]string{"deepsentinel-server.service", "deepsentinel-agent.service"}[d.component]
}

func (d systemdDaemon) serviceFilePath() string {
	return [...]string{"/etc/systemd/system/" + d.serviceFileName(), "/etc/systemd/system/" + d.serviceFileName()}[d.component]
}

func (d systemdDaemon) binaryPath() string {
	return [...]string{d.component.binaryPath(), d.component.binaryPath()}[d.component]
}

type systemdDaemon struct {
	component daemonType
}

func (d *systemdDaemon) isDaemonInstalled() bool {
	cmd := exec.Command("systemctl", "is-enabled", d.serviceFileName())
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) isDaemonRunning() bool {
	cmd := exec.Command("systemctl", "is-active", d.serviceFileName())
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) installDaemon() error {
	installBinary(d.component)

	// Copy the systemd-template.service to serviceFilePath
	template := `[Unit]
Description=<service>
After=network.target

[Service]
ExecStart=<binaryPath> run
Restart=always

[Install]
WantedBy=multi-user.target`
	template = strings.ReplaceAll(template, "<binaryPath>", d.component.binaryPath())
	template = strings.ReplaceAll(template, "<service>", "deepSentinel "+d.component.String())

	err := os.WriteFile(d.serviceFilePath(), []byte(template), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "enable", d.serviceFileName())
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (d *systemdDaemon) uninstallDaemon() error {
	err := d.stopDaemon()
	if err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "disable", d.serviceFileName())
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(d.serviceFilePath())
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(d.binaryPath())
	if err != nil {
		return err
	}

	return nil
}

func (d *systemdDaemon) updateDaemon() error {
	d.stopDaemon()

	err := os.Remove(d.binaryPath())
	if err != nil {
		return err
	}

	copyBinary(d.binaryPath())

	d.startDaemon()

	return nil
}

func (d *systemdDaemon) stopDaemon() error {
	err := agent.ExecuteConfigInstruction("unregister", nil)
	if err != nil {
		return err
	}

	// Wait for the agent to stop
	fmt.Println("Waiting for the agent to unregister successfully...")
	time.Sleep(5 * time.Second)

	cmd := exec.Command("systemctl", "stop", d.serviceFileName())
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) startDaemon() error {
	cmd := exec.Command("systemctl", "start", d.serviceFileName())
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) restartDaemon() error {
	cmd := exec.Command("systemctl", "restart", d.serviceFileName())
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
