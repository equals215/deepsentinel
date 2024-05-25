package monitoring

import (
	"errors"
	"sync"
	"time"

	"github.com/equals215/deepsentinel/alerting"
	"github.com/equals215/deepsentinel/config"
	log "github.com/sirupsen/logrus"
)

type statusType int

const (
	pass statusType = iota
	warn
	fail
)

func stringtoStatusType(str string) (statusType, error) {
	strStatus := map[string]statusType{
		"pass": pass,
		"warn": warn,
		"fail": fail,
	}

	if status, ok := strStatus[str]; ok {
		return status, nil
	}
	return fail, errors.New("invalid status string")
}

func (s statusType) String() string {
	statusStr := map[statusType]string{
		pass: "pass",
		warn: "warn",
		fail: "fail",
	}
	return statusStr[s]
}

type serviceStatus struct {
	status statusType
	count  int
}

type servicesStatus map[string]*serviceStatus

type timeSerieNode struct {
	timestamp time.Time
	services  servicesStatus
	previous  *timeSerieNode
}

// probeTimeSerie is a mutexed linked list of timeSerieNode
// It is used to store the status of the services
// The head of the list is the latest status
// It goes like : previous <- ... <- head
type probeTimeSerie struct {
	sync.Mutex
	head *timeSerieNode
	size int
}

func (p *probeWorker) workServices(payload *Payload) {
	p.timeSerie.Lock()
	p.storePayload(payload)
	if p.timeSerie.size > trimTimeSeriesThreshold {
		go p.trimTimeSerie()
	}
	p.checkAlert()
	p.timeSerie.Unlock()
}

func (p *probeWorker) storePayload(payload *Payload) {
	tempServiceStatus := make(servicesStatus)

	for service, state := range payload.Services {
		tempCount := 0
		parsedStatus, err := stringtoStatusType(state)
		if err != nil {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": payload.Machine,
				"service": service,
				"status":  p.status,
			}).Error("Invalid status string in payload, defaulting to fail")
		}

		if p.timeSerie.head != nil {
			prevServiceStatus, ok := p.timeSerie.head.services[service]
			if ok && prevServiceStatus.status == parsedStatus && prevServiceStatus.status != pass {
				tempCount = prevServiceStatus.count + 1
			}
		}

		tempServiceStatus[service] = &serviceStatus{
			status: parsedStatus,
			count:  tempCount,
		}

		log.WithFields(log.Fields{
			"probe":   p.name,
			"machine": payload.Machine,
			"service": service,
			"status":  parsedStatus.String(),
			"count":   tempCount,
		}).Trace("Service status stored in timeserie")
	}

	newNode := &timeSerieNode{
		timestamp: payload.Timestamp,
		services:  tempServiceStatus,
		previous:  p.timeSerie.head,
	}

	p.timeSerie.head = newNode
	p.timeSerie.size++

	log.WithFields(log.Fields{
		"probe":   p.name,
		"machine": payload.Machine,
		"status":  p.status,
		"size":    p.timeSerie.size,
	}).Trace("Payload stored in timeserie")
}

func (p *probeWorker) checkAlert() {
	if p.timeSerie.head == nil {
		return
	}

	for service, status := range p.timeSerie.head.services {
		var alertingStatus string
		toAlert := false

		lowThreshhold := config.Server.FailedToAlertedLowThreshold
		highThreshhold := config.Server.FailedToAlertedLowThreshold + config.Server.AlertedLowToAlertedHighThreshold

		if status.status == fail && status.count == lowThreshhold && status.count < highThreshhold {
			alertingStatus = "low"
			toAlert = true
		} else if status.status == fail && status.count == highThreshhold {
			alertingStatus = "high"
			toAlert = true
		} else if status.status == fail {
			toAlert = false
		}

		if toAlert {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": p.name,
				"service": service,
				"status":  fail,
			}).Warnf("Service in fail status. Alerting %s", alertingStatus)
			service = p.name + "-" + service
			alerting.ServerAlert("service", service, alertingStatus)
		} else if status.status == fail && status.count > lowThreshhold && status.count%10 == 0 {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": p.name,
				"service": service,
				"status":  fail,
			}).Warn("Service still in fail status. Alerady alerted")
		}
	}
}
