package monitoring

import (
	"encoding/json"
	"time"

	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
)

type probeStatus int

const (
	normal probeStatus = iota
	degraded
	failed
	alertedLow
	alertedHigh
)

type payload struct {
	MachineStatus string            `json:"machineStatus"`
	Services      map[string]string `json:"services"`
}

type probeObject struct {
	name          string
	c             chan payload
	status        probeStatus
	counter       int
	lastNormal    time.Time
	timeSerieHead *timeSerieNode
}

var probes = make(map[string]probeObject)

// IngestPayload function handles the payload from the API server
func IngestPayload(machine string, rawPayload []byte) error {
	parsedPayload := payload{}
	err := json.Unmarshal(rawPayload, &parsedPayload)
	if err != nil {
		log.Errorf("Error unmarshalling payload: %s\n", err)
		return err
	}

	if probe, ok := probes[machine]; ok {
		probe.c <- parsedPayload
	} else {
		probe := probeObject{name: machine, c: make(chan payload), status: normal, counter: 0}
		probes[machine] = probe
		log.Infof("Starting probe for machine: %s\n", machine)
		go probe.work()
		probe.c <- parsedPayload
	}
	return nil
}

func (p *probeObject) work() {
	inactivityDelay := time.Duration(config.Server.ProbeInactivityDelaySeconds) * time.Second
	timer := time.NewTimer(inactivityDelay)
	for {
		select {
		case payload := <-p.c:
			log.Infof("Received payload: %+v\n", payload)
			p.reset()
			timer.Reset(inactivityDelay)
		case <-timer.C:
			p.timerIncrement()
			timer.Reset(inactivityDelay)
		}
	}
}

func (p *probeObject) timerIncrement() {
	switch p.status {
	case normal:
		p.updateStatus(degraded)
	case degraded:
		p.counter++
		if p.counter >= config.Server.DegradedToFailedThreshold {
			p.updateStatus(failed)
			break
		}
	case failed:
		p.counter++
		if p.counter >= config.Server.FailedToAlertedLowThreshold {
			p.updateStatus(alertedLow)
			break
		}
	case alertedLow:
		p.counter++
		if p.counter >= config.Server.AlertedLowToAlertedHighThreshold {
			p.updateStatus(alertedHigh)
			break
		}
	case alertedHigh:
		p.counter++
		if p.counter%10 == 0 {
			log.Warnf("Machine %s is still in %s state\n", p.name, "alertedHigh")
		}
	}
}

func (p *probeObject) reset() {
	if p.status > normal {
		log.Infof("Machine %s is back in normal state\n", p.name)
	}
	p.status = normal
	p.counter = 0
	p.lastNormal = time.Now()
}

func (p *probeObject) updateStatus(status probeStatus) {
	statusStr := map[probeStatus]string{
		normal:      "normal",
		degraded:    "degraded",
		failed:      "failed",
		alertedLow:  "alertedLow",
		alertedHigh: "alertedHigh",
	}

	p.status = status
	p.counter = 0
	if status > normal {
		duration := time.Since(p.lastNormal)
		log.Warnf("No payload received for %s. Machine %s is now in %s state\n", duration.String(), p.name, statusStr[status])
		return
	}
	log.Infof("Machine %s is now in %s state\n", p.name, statusStr[status])
}
