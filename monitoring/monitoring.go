// Package monitoring provides the logic to monitor the services and machines.
package monitoring

import (
	"strings"
	"sync"
	"time"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/dashboard"
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
	sync.Mutex
	name       string
	data       chan *Payload
	stop       chan bool
	status     probeStatus
	counter    int
	lastNormal time.Time
	timeSerie  *probeTimeSerie
}

// Handle function handles the payload from the API server
func Handle(channel chan *Payload, dashboardOperator *dashboard.Operator) {
	log.Debug("Starting monitoring.Handle")
	var probeMap sync.Map
	var probeList = make([]string, 0)
	var timer = time.NewTimer(5 * time.Second)

	for {
		select {
		case <-timer.C:
			if dashboardOperator == nil {
				continue
			}

			dashboardPayload := &dashboard.Data{
				Probes: make([]*dashboard.Probe, 0),
			}

			for _, probe := range probeList {
				if loaded, ok := probeMap.Load(probe); ok {
					probe := loaded.(*probeObject)
					probe.Lock()
					dashboardProbe := &dashboard.Probe{
						Name:   strings.Clone(probe.name),
						Status: strings.Clone(probe.status.String()),
					}
					probe.Unlock()
					dashboardPayload.Probes = append(dashboardPayload.Probes, dashboardProbe)
				}
			}

			dashboardOperator.In <- dashboardPayload
			timer.Reset(5 * time.Second)
			continue
		case payload := <-channel:
			if loaded, ok := probeMap.Load(payload.Machine); ok {
				probe := loaded.(*probeObject)
				if payload.MachineStatus == "delete" {
					// Delete the probe
					probe.Lock()
					for i, name := range probeList {
						if name == payload.Machine {
							probeList = append(probeList[:i], probeList[i+1:]...)
							break
						}
					}
					probe.delete()
					probeMap.Delete(payload.Machine)
				} else {
					// Send the payload to the probe
					probe.data <- payload
				}
			} else {
				// Create a new probe
				newProbe := makeProbe(payload)

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

				probeList = append(probeList, payload.Machine)
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
			p.Lock()
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
			p.Unlock()
		case <-timer.C:
			p.Lock()
			p.timerIncrement()
			timer.Reset(inactivityDelay)
			p.Unlock()
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

func (p *probeObject) delete() {
	log.WithFields(log.Fields{
		"probe": p.name,
	}).Info("Deleting probe")
	p.stop <- true
	close(p.data)
	close(p.stop)
}

func makeProbe(originPayload *Payload) *probeObject {
	return &probeObject{
		name:       originPayload.Machine,
		data:       make(chan *Payload, 1),
		stop:       make(chan bool),
		status:     normal,
		counter:    0,
		lastNormal: time.Now(),
		timeSerie: &probeTimeSerie{
			head: &timeSerieNode{
				timestamp: originPayload.Timestamp,
				services:  make(map[string]*serviceStatus),
				previous:  nil,
			},
			size: 1,
		},
	}
}
