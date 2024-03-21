package monitoring

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type timeSerieNode struct {
	timestamp time.Time
	services  map[string]string
	previous  *timeSerieNode
}

// StoreStatus function : stores the payload into the timeserie
func (p *probeObject) StorePayload(payload *Payload) {
	newNode := &timeSerieNode{
		timestamp: payload.Timestamp,
		services:  payload.Services,
		previous:  p.timeSerieHead,
	}
	p.timeSerieHead = newNode
	log.WithFields(log.Fields{
		"probe":   p.name,
		"machine": payload.Machine,
		"status":  p.status,
	}).Trace("Payload stored in timeserie")
}

// GetStatus function : calculate the diff between the last and the current payload
