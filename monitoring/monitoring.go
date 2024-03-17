package monitoring

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type payload struct {
	MachineStatus string            `json:"machineStatus"`
	Services      map[string]string `json:"services"`
}

type probeObject struct {
	c chan payload
}

var probes = make(map[string]probeObject)

func IngestPayload(machine string, rawPayload []byte) error {
	parsedPayload := payload{}
	err := json.Unmarshal(rawPayload, &parsedPayload)
	if err != nil {
		log.Errorf("Error unmarshalling payload: %s\n", err)
		return err
	}

	if probe, ok := probes[machine]; ok {
		probe.c <- parsedPayload
	} else {
		probe := probeObject{c: make(chan payload)}
		probes[machine] = probe
		log.Infof("Starting probe for machine: %s\n", machine)
		go probe.work()
		probe.c <- parsedPayload
	}
	return nil
}

func (p *probeObject) work() {
	for {
		payload := <-p.c
		log.Infof("Received payload: %+v\n", payload)
	}
}
