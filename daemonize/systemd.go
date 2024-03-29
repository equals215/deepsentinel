package daemonize

import (
	"os"
	"os/exec"
	"strings"
)

var (
	serverServiceFileName = "deepsentinel-server.service"
	clientServiceFileName = "deepsentinel-client.service"
	serverServiceFilePath = "/etc/systemd/system/" + serverServiceFileName
	clientServiceFilePath = "/etc/systemd/system/" + clientServiceFileName
)

type systemdDaemon struct {
	daemonFilePath string
}

func (d *systemdDaemon) isDaemonInstalled(component daemonType) bool {
	var serviceFileName string
	if component == Server {
		serviceFileName = serverServiceFileName
	} else if component == Client {
		serviceFileName = clientServiceFileName
	}

	cmd := exec.Command("systemctl", "is-enabled", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) isDaemonRunning(component daemonType) bool {
	var serviceFileName string
	if component == Server {
		serviceFileName = serverServiceFileName
	} else if component == Client {
		serviceFileName = clientServiceFileName
	}

	cmd := exec.Command("systemctl", "is-active", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func (d *systemdDaemon) installDaemon(component daemonType) error {
	var serviceFileName string
	var binaryPath string
	var serviceFilePath string
	var serviceName string
	if component == Server {
		serviceFileName = serverServiceFileName
		binaryPath = serverBinaryPath
		serviceFilePath = serverServiceFilePath
		serviceName = "deepSentinel Server"
	} else if component == Client {
		serviceFileName = clientServiceFileName
		binaryPath = clientBinaryPath
		serviceFilePath = clientServiceFilePath
		serviceName = "deepSentinel Client"
	}

	// Copy the systemd-template.service to serviceFilePath
	template := `[Unit]
Description=<service>
After=network.target

[Service]
ExecStart=<binaryPath>
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

func (d *systemdDaemon) uninstallDaemon(component daemonType) error {
	var serviceFileName string
	var serviceFilePath string
	var binaryPath string
	if component == Server {
		serviceFileName = serverServiceFileName
		serviceFilePath = serverServiceFilePath
		binaryPath = serverBinaryPath
	} else if component == Client {
		serviceFileName = clientServiceFileName
		serviceFilePath = clientServiceFilePath
		binaryPath = clientBinaryPath
	}

	err := d.stopDaemon(component)
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

func (d *systemdDaemon) updateDaemon(component daemonType) error {
	var binaryPath string
	if component == Server {
		binaryPath = serverBinaryPath
	} else if component == Client {
		binaryPath = clientBinaryPath
	}

	d.stopDaemon(component)

	err := os.Remove(binaryPath)
	if err != nil {
		return err
	}

	copyBinary(binaryPath)

	d.startDaemon(component)

	return nil
}

func (d *systemdDaemon) stopDaemon(component daemonType) error {
	var serviceFileName string
	if component == Server {
		serviceFileName = serverServiceFileName
	} else if component == Client {
		serviceFileName = clientServiceFileName
	}

	cmd := exec.Command("systemctl", "stop", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) startDaemon(component daemonType) error {
	var serviceFileName string
	if component == Server {
		serviceFileName = serverServiceFileName
	} else if component == Client {
		serviceFileName = clientServiceFileName
	}

	cmd := exec.Command("systemctl", "start", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *systemdDaemon) restartDaemon(component daemonType) error {
	var serviceFileName string
	if component == Server {
		serviceFileName = serverServiceFileName
	} else if component == Client {
		serviceFileName = clientServiceFileName
	}

	cmd := exec.Command("systemctl", "restart", serviceFileName)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
