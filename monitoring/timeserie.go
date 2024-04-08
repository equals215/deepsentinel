package monitoring

import (
	"errors"
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

const trimTimeSeriesThreshold = 100

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

type serviceStatus struct {
	status statusType
	count  int
}

type timeSerieNode struct {
	timestamp time.Time
	services  map[string]*serviceStatus
	previous  *timeSerieNode
}

func (p *probeObject) workServices(payload *Payload) {
	p.timeSerieMutex.Lock()
	p.storePayload(payload)
	if p.timeSerieSize > trimTimeSeriesThreshold {
		go p.trimTimeSerie()
	}
	p.checkAlert()
	p.timeSerieMutex.Unlock()
}

func (p *probeObject) storePayload(payload *Payload) {
	tempServiceStatus := make(map[string]*serviceStatus)
	tempCount := 0

	for service, state := range payload.Services {
		parsedStatus, err := stringtoStatusType(state)

		if err != nil {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": payload.Machine,
				"service": service,
				"status":  p.status,
			}).Error("Invalid status string in payload, defaulting to fail")
		}

		if p.timeSerieHead != nil {
			prevServiceStatus, ok := p.timeSerieHead.services[service]
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
			"status":  parsedStatus,
			"count":   tempCount,
		}).Trace("Service status stored in timeserie")
	}

	newNode := &timeSerieNode{
		timestamp: payload.Timestamp,
		services:  tempServiceStatus,
		previous:  p.timeSerieHead,
	}

	p.timeSerieHead = newNode
	p.timeSerieSize++

	log.WithFields(log.Fields{
		"probe":   p.name,
		"machine": payload.Machine,
		"status":  p.status,
		"size":    p.timeSerieSize,
	}).Trace("Payload stored in timeserie")
}

func (p *probeObject) trimTimeSerie() {
	p.timeSerieMutex.Lock()
	defer p.timeSerieMutex.Unlock()

	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerieSize,
	}).Trace("Trimming timeserie")

	count := 0
	historicalAllPass := true
	currentNode := p.timeSerieHead
	for currentNode != nil {
		for _, service := range currentNode.services {
			if service.status != pass {
				historicalAllPass = false
				break
			}
		}
		if !historicalAllPass {
			break
		}
		if count >= trimTimeSeriesThreshold {
			break
		}
		count++
		currentNode = currentNode.previous
	}

	if historicalAllPass {
		p.timeSerieHead.previous = nil
		p.timeSerieSize = 1
		log.WithFields(log.Fields{
			"probe": p.name,
			"size":  p.timeSerieSize,
		}).Trace("All historical data is pass trimmed timeserie to latest node")
		return
	}

	if count >= trimTimeSeriesThreshold {
		currentNode = p.timeSerieHead
		for i := 0; i < trimTimeSeriesThreshold-1; i++ {
			currentNode = currentNode.previous
		}
		currentNode.previous = nil
		p.timeSerieSize = trimTimeSeriesThreshold
		log.WithFields(log.Fields{
			"probe": p.name,
			"size":  p.timeSerieSize,
		}).Trace("Trimmed timeserie to last 10 nodes")
		return
	}

	if currentNode != nil {
		currentNode.previous = nil
		p.timeSerieSize = count
		log.WithFields(log.Fields{
			"probe": p.name,
			"size":  p.timeSerieSize,
		}).Trace("Trimmed timeserie to node with first non-pass status")
		return
	}

	p.timeSerieSize = 0
	p.timeSerieHead = nil
	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerieSize,
	}).Error("Trimmed timeserie to empty")
}

func (p *probeObject) checkAlert() {
	if p.timeSerieHead == nil {
		return
	}

	for service, status := range p.timeSerieHead.services {
		var alertingStatus string
		alert := false

		lowThreshhold := config.Server.FailedToAlertedLowThreshold
		highThreshhold := config.Server.FailedToAlertedLowThreshold + config.Server.AlertedLowToAlertedHighThreshold

		if status.status == fail && status.count >= lowThreshhold && status.count < highThreshhold {
			alertingStatus = "low"
			alert = true
		} else if status.status == fail && status.count == highThreshhold {
			alertingStatus = "high"
			alert = true
		} else if status.status == fail && status.count > highThreshhold {
			alert = false
		} else {
			continue
		}

		if alert {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": p.name,
				"service": service,
				"status":  fail,
			}).Warnf("Service in fail status. Alerting %s", alertingStatus)
			service = p.name + "-" + service
			alerting.ServerAlert("service", service, alertingStatus)
		} else if !alert && status.count%10 == 0 {
			log.WithFields(log.Fields{
				"probe":   p.name,
				"machine": p.name,
				"service": service,
				"status":  fail,
			}).Warn("Service still in fail status. Alerady alerted")
		}
	}
}
