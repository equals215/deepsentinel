package dashboard

import (
	"sync"
	"time"
)

type Probe struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Data struct {
	Probes []*Probe `json:"probes"`
}

type Operator struct {
	sync.Mutex
	workers     map[int]chan *Data
	PollingFreq time.Duration
	In          chan *Data
}

func Handle() *Operator {
	operator := &Operator{
		workers:     make(map[int]chan *Data),
		PollingFreq: 5 * time.Second,
		In:          make(chan *Data),
	}
	go func(operator *Operator) {
		for {
			select {
			case data := <-operator.In:
				operator.Lock()
				for _, worker := range operator.workers {
					worker <- data
				}
				operator.Unlock()
			}
		}
	}(operator)
	return operator
}

func (o *Operator) NewWorker() (chan *Data, int) {
	worker := make(chan *Data)
	o.Lock()
	i := len(o.workers)
	o.workers[i] = worker
	o.Unlock()
	return worker, i
}

func (o *Operator) RemoveWorker(i int) {
	o.Lock()
	close(o.workers[i])
	delete(o.workers, i)
	o.Unlock()
}
