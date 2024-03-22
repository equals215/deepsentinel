package monitoring

import (
	"errors"
	"time"

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
	} else {
		return fail, errors.New("invalid status string")
	}
}

type serviceStatus struct {
	status statusType
}

type timeSerieNode struct {
	timestamp time.Time
	services  *map[string]*serviceStatus
	previous  *timeSerieNode
}

// StorePayloads function : store the payload in the timeserie
func (p *probeObject) StorePayload(payload *Payload) {
	tempServiceStatus := make(map[string]*serviceStatus)
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
		tempServiceStatus[service] = &serviceStatus{
			status: parsedStatus,
		}
	}

	newNode := &timeSerieNode{
		timestamp: payload.Timestamp,
		services:  &tempServiceStatus,
		previous:  p.timeSerieHead,
	}

	p.timeSerieHead = newNode

	log.WithFields(log.Fields{
		"probe":   p.name,
		"machine": payload.Machine,
		"status":  p.status,
	}).Trace("Payload stored in timeserie")
}
