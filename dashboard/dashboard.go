package dashboard

import (
	"sync"
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
	workers map[int]chan *Data
}

func Handle(dataChan chan *Data) *Operator {
	operator := &Operator{
		workers: make(map[int]chan *Data),
	}
	go func() {
		for {
			select {
			case data := <-dataChan:
				operator.Lock()
				for _, worker := range operator.workers {
					worker <- data
				}
				operator.Unlock()
			}
		}
	}()
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
