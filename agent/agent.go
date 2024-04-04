package agent

import (
	"time"

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
		err := reportAlive()
		if err != nil {
			log.Errorf("error reporting alive: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}
