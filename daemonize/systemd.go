package daemonize

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/equals215/deepsentinel/agent"
)

var (
	systemdServerServiceFileName = "deepsentinel-server.service"
	systemdAgentServiceFileName  = "deepsentinel-agent.service"
	systemdServerServiceFilePath = "/etc/systemd/system/" + systemdServerServiceFileName
	systemdAgentServiceFilePath  = "/etc/systemd/system/" + systemdAgentServiceFileName
)

type systemdDaemon struct {
	component daemonType
}

func (d *systemdDaemon) isDaemonInstalled() bool {
	var serviceFileName string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
	}

	cmd := exec.Command("systemctl", "is-enabled", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) isDaemonRunning() bool {
	var serviceFileName string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
	}

	cmd := exec.Command("systemctl", "is-active", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) installDaemon() error {
	var serviceFileName string
	var binaryPath string
	var serviceFilePath string
	var serviceName string

	installBinary(d.component)

	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
		binaryPath = serverBinaryPath
		serviceFilePath = systemdServerServiceFilePath
		serviceName = "deepSentinel Server"
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
		binaryPath = agentBinaryPath
		serviceFilePath = systemdAgentServiceFilePath
		serviceName = "deepSentinel Agent"
	}

	// Copy the systemd-template.service to serviceFilePath
	template := `[Unit]
Description=<service>
After=network.target

[Service]
ExecStart=<binaryPath> run
Restart=always

[Install]
WantedBy=multi-user.target`
	template = strings.ReplaceAll(template, "<binaryPath>", binaryPath)
	template = strings.ReplaceAll(template, "<service>", serviceName)

	err := os.WriteFile(serviceFilePath, []byte(template), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "enable", serviceFileName)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (d *systemdDaemon) uninstallDaemon() error {
	var serviceFileName string
	var serviceFilePath string
	var binaryPath string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
		serviceFilePath = systemdServerServiceFilePath
		binaryPath = serverBinaryPath
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
		serviceFilePath = systemdAgentServiceFilePath
		binaryPath = agentBinaryPath
	}

	err := d.stopDaemon()
	if err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "disable", serviceFileName)
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(serviceFilePath)
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(binaryPath)
	if err != nil {
		return err
	}

	return nil
}

func (d *systemdDaemon) updateDaemon() error {
	var binaryPath string
	if d.component == Server {
		binaryPath = serverBinaryPath
	} else if d.component == Agent {
		binaryPath = agentBinaryPath
	}

	d.stopDaemon()

	err := os.Remove(binaryPath)
	if err != nil {
		return err
	}

	copyBinary(binaryPath)

	d.startDaemon()

	return nil
}

func (d *systemdDaemon) stopDaemon() error {
	var serviceFileName string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
	}

	err := agent.DoConfigInstruction("unregister", nil)
	if err != nil {
		return err
	}

	// Wait for the agent to stop
	fmt.Println("Waiting for the agent to unregister successfully...")
	time.Sleep(5 * time.Second)

	cmd := exec.Command("systemctl", "stop", serviceFileName)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) startDaemon() error {
	var serviceFileName string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
	}

	cmd := exec.Command("systemctl", "start", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) restartDaemon() error {
	var serviceFileName string
	if d.component == Server {
		serviceFileName = systemdServerServiceFileName
	} else if d.component == Agent {
		serviceFileName = systemdAgentServiceFileName
	}

	cmd := exec.Command("systemctl", "restart", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
