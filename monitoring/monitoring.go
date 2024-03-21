package monitoring

import (
	"sync"
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

// Payload is the structure of the payload received from the API server
type Payload struct {
	MachineStatus string            `json:"machineStatus"`
	Services      map[string]string `json:"services"`
	Timestamp     time.Time
	Machine       string
}

type probeObject struct {
	name          string
	c             chan *Payload
	status        probeStatus
	counter       int
	lastNormal    time.Time
	timeSerieHead *timeSerieNode
}

// Handle function handles the payload from the API server
func Handle(channel chan *Payload) {
	log.Debug("Starting monitoring.Handle")
	var probes struct {
		sync.Mutex
		p map[string]*probeObject
	}
	probes.p = make(map[string]*probeObject)
	for {
		select {
		case receivedPayload := <-channel:
			probes.Lock()
			if probe, ok := probes.p[receivedPayload.Machine]; ok {
				probe.c <- receivedPayload
			} else {
				probe := &probeObject{
					name:    receivedPayload.Machine,
					c:       make(chan *Payload),
					status:  normal,
					counter: 0,
				}

				probes.p[receivedPayload.Machine] = probe
				log.WithFields(log.Fields{
					"probe":   probe.name,
					"machine": receivedPayload.Machine,
					"status":  probe.status,
				}).Info("Starting probe thread")
				go probe.work()
				probe.c <- receivedPayload
			}
			probes.Unlock()
		}
	}
}

func (p *probeObject) work() {
	inactivityDelay := time.Duration(config.Server.ProbeInactivityDelaySeconds) * time.Second
	timer := time.NewTimer(inactivityDelay)
	for {
		select {
		case payload := <-p.c:
			if payload.Machine != p.name {
				log.WithFields(log.Fields{
					"probe":   p.name,
					"machine": payload.Machine,
				}).Fatal("Discrepancy between probe and machine")
			}
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": payload.Machine,
				"status":  p.status,
			}).Info("Received report")
			p.StorePayload(payload)
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

	p.status++
	p.counter = 0
	if status > normal {
		duration := time.Since(p.lastNormal)
		log.Warnf("No payload received for %s. Machine %s is now in %s state\n", duration.String(), p.name, statusStr[status])
		return
	}
	log.Infof("Machine %s is now in %s state\n", p.name, statusStr[status])
}
