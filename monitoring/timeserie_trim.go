package monitoring

import log "github.com/sirupsen/logrus"

const trimTimeSeriesThreshold = 100

func (p *probeObject) trimTimeSerie() {
	p.timeSerie.Lock()
	defer p.timeSerie.Unlock()

	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerie.size,
	}).Trace("Trimming timeserie")

	count := 0
	historicalAllPass := true
	currentNode := p.timeSerie.head
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
		// If all historical data is pass, keep only the latest node
		p.trimToLastNode()
	} else if count >= trimTimeSeriesThreshold {
		// If historical data is not all pass and bigger than trimTimeSeriesThreshold
		// trim to the last 10 nodes
		p.trimToNode(trimTimeSeriesThreshold)
	} else if currentNode != nil {
		// If historical data is not all pass
		// trim to the first non-pass node
		p.trimToGivenNode(currentNode, count)
	} else {
		// Else trim to empty
		p.timeSerie.size = 0
		p.timeSerie.head = nil
		log.WithFields(log.Fields{
			"probe": p.name,
			"size":  p.timeSerie.size,
		}).Error("Trimmed timeserie to empty")
	}
}

func (p *probeObject) trimToLastNode() {
	p.timeSerie.head.previous = nil
	p.timeSerie.size = 1
	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerie.size,
	}).Trace("All historical data is pass trimmed timeserie to latest node")
	return
}

func (p *probeObject) trimToNode(count int) {
	currentNode := p.timeSerie.head
	for i := 0; i < count-1; i++ {
		currentNode = currentNode.previous
	}
	currentNode.previous = nil
	p.timeSerie.size = trimTimeSeriesThreshold
	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerie.size,
	}).Trace("Trimmed timeserie to last 10 nodes")
	return
}

func (p *probeObject) trimToGivenNode(currentNode *timeSerieNode, count int) {
	currentNode.previous = nil
	p.timeSerie.size = count
	log.WithFields(log.Fields{
		"probe": p.name,
		"size":  p.timeSerie.size,
	}).Trace("Trimmed timeserie to node with first non-pass status")
	return
}
