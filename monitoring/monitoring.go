package monitoring

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
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

type probeList struct {
	sync.Mutex
	p           map[string]*probeObject
	probesNames *[]string
}

func CreateProbeList() *probeList {
	probesNames := make([]string, 0)
	return &probeList{
		p:           make(map[string]*probeObject),
		probesNames: &probesNames,
	}
}

// Handle function handles the payload from the API server
func (probes *probeList) Handle(channel chan Payload) {
	log.Debug("Starting monitoring.Handle")
	// if probes != nil {
	// 	log.Fatal("probes already initialized, this should never happen, please open an issue https://github.com/equals215/deepsentinel/issues/new")
	// }
	for {
		select {
		case receivedPayload := <-channel:
			probes.Lock()
			log.Infof("probesNames: %v\n", probes.probesNames)
			spew.Dump(probes)
			payload := &receivedPayload
			spew.Dump(payload)
			// ONLY FOR TESTING SHOULD BE DELETED
			// ONLY FOR TESTING SHOULD BE DELETED
			if len(probes.p) > 2 {
				log.WithFields(log.Fields{
					"probes": probes.p,
				}).Info("Probes")
				spew.Dump(probes)
				for _, probeName := range *probes.probesNames {
					log.WithFields(log.Fields{
						"probe": probeName,
						"hexa":  fmt.Sprintf("%x", probeName),
						"b64":   base64.StdEncoding.EncodeToString([]byte(probeName)),
					}).Info("Probes")
				}
				log.Fatalf("Too many probes: %d", len(probes.p))
			}
			// ONLY FOR TESTING SHOULD BE DELETED
			// ONLY FOR TESTING SHOULD BE DELETED
			if probe, ok := probes.p[payload.Machine]; ok {
				if payload.MachineStatus == "delete" {
					// Delete the probe
					log.WithFields(log.Fields{
						"probe":   probe.name,
						"machine": payload.Machine,
					}).Info("Deleting probe")
					probe.stop <- true
					close(probe.data)
					close(probe.stop)
					delete(probes.p, payload.Machine)
					probes.Unlock()
					log.Info("—————————————————————————————————————————————————————————————————————")
					continue
				}
				// Send the payload to the probe
				// probe.data <- payload
				probes.Unlock()
				log.Info("—————————————————————————————————————————————————————————————————————")
				continue
			} else {
				// ONLY FOR TESTING SHOULD BE DELETED
				// ONLY FOR TESTING SHOULD BE DELETED
				for _, probeName := range *probes.probesNames {
					log.Infof("payload.machine=%s, probeName=%s\n", payload.Machine, probeName)
					if payload.Machine == probeName {
						log.WithFields(log.Fields{
							"gotMachine":      payload.Machine,
							"existingMachine": probeName,
						}).Fatalf("Probes")
					}
				}
				spew.Dump(payload)
				// ONLY FOR TESTING SHOULD BE DELETED
				// ONLY FOR TESTING SHOULD BE DELETED
				// Create a new probe
				probe := &probeObject{
					name:       payload.Machine,
					data:       make(chan *Payload, 1),
					stop:       make(chan bool),
					status:     normal,
					counter:    0,
					lastNormal: time.Now(),
					timeSerieHead: &timeSerieNode{
						timestamp: payload.Timestamp,
						services:  make(map[string]*serviceStatus),
						previous:  nil,
					},
					timeSerieSize:  1,
					timeSerieMutex: sync.Mutex{},
				}

				probes.p[payload.Machine] = probe
				newProbesNames := append(*probes.probesNames, payload.Machine)
				probes.probesNames = &newProbesNames
				log.WithFields(log.Fields{
					"probe":   probe.name,
					"machine": payload.Machine,
					"status":  probe.status,
				}).Info("Starting probe thread")
				go probe.work()
				// probe.data <- payload
				probes.Unlock()
				log.Info("—————————————————————————————————————————————————————————————————————")
				continue
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
