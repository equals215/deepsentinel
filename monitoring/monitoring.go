package monitoring

import (
	"sync"
	"time"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config/v1"
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

func (s probeStatus) String() string {
	statusStr := map[probeStatus]string{
		normal:      "normal",
		degraded:    "degraded",
		failed:      "failed",
		alertedLow:  "alertedLow",
		alertedHigh: "alertedHigh",
	}
	return statusStr[s]
}

// Payload is the structure of the payload received from the API server
type Payload struct {
	MachineStatus string            `json:"machineStatus"`
	Services      map[string]string `json:"services"`
	Timestamp     time.Time
	Machine       string
}

type probeObject struct {
	name           string
	data           chan *Payload
	stop           chan bool
	status         probeStatus
	counter        int
	lastNormal     time.Time
	timeSerieHead  *timeSerieNode
	timeSerieSize  int
	timeSerieMutex sync.Mutex
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
				if receivedPayload.MachineStatus == "delete" {
					log.WithFields(log.Fields{
						"probe":   probe.name,
						"machine": receivedPayload.Machine,
					}).Info("Deleting probe")
					probe.stop <- true
					close(probe.data)
					close(probe.stop)
					delete(probes.p, receivedPayload.Machine)
				} else {
					probe.data <- receivedPayload
				}
			} else {
				probe := &probeObject{
					name:    receivedPayload.Machine,
					data:    make(chan *Payload),
					stop:    make(chan bool),
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
				probe.data <- receivedPayload
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
		case <-p.stop:
			return
		case payload := <-p.data:
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
			}).Trace("Received report")
			p.workServices(payload)
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
		p.updateStatus()
	case degraded:
		p.counter++
		if p.counter >= config.Server.DegradedToFailedThreshold {
			p.updateStatus()
			break
		}
	case failed:
		p.counter++
		if p.counter >= config.Server.FailedToAlertedLowThreshold {
			p.updateStatus()
			alerting.ServerAlert("machine", p.name, "low")
			break
		}
	case alertedLow:
		p.counter++
		if p.counter >= config.Server.AlertedLowToAlertedHighThreshold {
			p.updateStatus()
			alerting.ServerAlert("machine", p.name, "high")
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

func (p *probeObject) updateStatus() {
	p.status++
	p.counter = 0
	if p.status > normal {
		duration := time.Since(p.lastNormal)
		log.Warnf("No payload received for %s. Machine %s is now in %s state\n", duration.String(), p.name, p.status.String())
		return
	}
	log.Infof("Machine %s is now in %s state\n", p.name, p.status.String())
}
