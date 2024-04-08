// Package agent defines agent functionnality
package agent

import (
	"time"

	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
)

func work() {
	for {
		stop.Lock()
		if stop.val {
			stop.Unlock()
			err := reportUnregisterAgent()
			if err != nil {
				log.Errorf("error unregistering agent: %v", err)
			}
			return
		}
		stop.Unlock()
		if config.Agent.ServerAddress == "" || config.Agent.AuthToken == "" || config.Agent.MachineName == "" {
			log.Error("missing mandatory configuration, please run deepsentinel config server-address, auth-token, and machine-name")
			stop.Lock()
			stop.val = true
			stop.Unlock()
			return
		}
		err := reportAlive()
		if err != nil {
			log.Errorf("error reporting alive: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}
