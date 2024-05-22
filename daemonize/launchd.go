package daemonize

import (
	"fmt"
	"time"

	"github.com/brasic/launchd"
	"github.com/brasic/launchd/state"
	"github.com/equals215/deepsentinel/agent"
)

var (
	launchdServerServiceFileName = "com.deepsentinel.server.plist"
	launchdAgentServiceFileName  = "com.deepsentinel.agent.plist"
	launchdServerServiceFilePath = "/Library/LaunchDaemons/" + launchdServerServiceFileName
	launchdAgentServiceFilePath  = "/Library/LaunchDaemons/" + launchdAgentServiceFileName
)

type launchdDaemon struct {
	component daemonType
	service   launchd.Service
}

func (d *launchdDaemon) primeDaemon() {
	d.service.Name = fmt.Sprintf("%s-%s", "deepsentinel", d.component.String())
	d.service.ExecutablePath = d.component.binaryPath()
	d.service.Argv = []string{"run"}
	d.service.RunAtLoad = true
	d.service.KeepAlive = true
}

func (d *launchdDaemon) isDaemonInstalled() bool {
	daemonState := d.service.InstallState()
	return daemonState.Is(state.Installed)
}

func (d *launchdDaemon) isDaemonRunning() bool {
	return d.service.IsHealthy()
}

func (d *launchdDaemon) installDaemon() error {
	installBinary(d.component)
	return d.service.Install()
}

func (d *launchdDaemon) uninstallDaemon() error {
	removePlist := true

	err := agent.ExecuteConfigInstruction("unregister", nil)
	if err != nil {
		return err
	}

	// Wait for the agent to stop
	fmt.Println("Waiting for the agent to unregister successfully...")
	time.Sleep(5 * time.Second)

	return d.service.Bootout(removePlist)
}

func (d *launchdDaemon) updateDaemon() error {
	err := d.uninstallDaemon()
	if err != nil {
		return err
	}

	return d.installDaemon()
}

func (d *launchdDaemon) stopDaemon() error {
	err := agent.ExecuteConfigInstruction("unregister", nil)
	if err != nil {
		return err
	}

	return d.service.Stop()
}

func (d *launchdDaemon) startDaemon() error {
	return d.service.Start()
}

func (d *launchdDaemon) restartDaemon() error {
	err := d.stopDaemon()
	if err != nil {
		return err
	}
	return d.startDaemon()
}
