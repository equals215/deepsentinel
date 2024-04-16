package monitoring

import (
	"sync"
	"time"

	"github.com/equals215/deepsentinel/alerting"
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
	MachineStatus string            `json:"machineStatus,omitempty"`
	Services      map[string]string `json:"services"`
	Timestamp     time.Time         `json:"-"`
	Machine       string            `json:"-"`
}

type probeObject struct {
	name       string
	data       chan *Payload
	stop       chan bool
	status     probeStatus
	counter    int
	lastNormal time.Time
	timeSerie  *probeTimeSerie
}

// Handle function handles the payload from the API server
func Handle(channel chan *Payload) {
	log.Debug("Starting monitoring.Handle")
	var probeMap sync.Map
	for {
		select {
		case payload := <-channel:
			if loaded, ok := probeMap.Load(payload.Machine); ok {
				probe := loaded.(*probeObject)
				if payload.MachineStatus == "delete" {
					// Delete the probe
					log.WithFields(log.Fields{
						"probe":   probe.name,
						"machine": payload.Machine,
					}).Info("Deleting probe")
					probe.stop <- true
					close(probe.data)
					close(probe.stop)
					probeMap.Delete(payload.Machine)
				} else {
					// Send the payload to the probe
					probe.data <- payload
				}
			} else {
				// Create a new probe
				newProbe := &probeObject{
					name:       payload.Machine,
					data:       make(chan *Payload, 1),
					stop:       make(chan bool),
					status:     normal,
					counter:    0,
					lastNormal: time.Now(),
					timeSerie: &probeTimeSerie{
						head: &timeSerieNode{
							timestamp: payload.Timestamp,
							services:  make(map[string]*serviceStatus),
							previous:  nil,
						},
						size: 1,
					},
				}

				value, loaded := probeMap.LoadOrStore(payload.Machine, newProbe)
				if loaded {
					log.WithFields(log.Fields{
						"machine": payload.Machine,
					}).Fatal("Machine already exists")
				}

				probe, ok := value.(*probeObject)
				if !ok {
					log.WithFields(log.Fields{
						"machine": payload.Machine,
					}).Fatal("Failed to load probe")
				}

				log.WithFields(log.Fields{
					"probe":   probe.name,
					"machine": payload.Machine,
					"status":  probe.status,
				}).Info("Starting probe thread")

				go probe.work()
				probe.data <- payload
			}
		}
	}
}

func (p *probeObject) work() {
	inactivityDelay, err := time.ParseDuration(config.Server.ProbeInactivityDelay)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse inactivity delay")
	}
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
