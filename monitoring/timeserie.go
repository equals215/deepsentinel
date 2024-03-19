package monitoring

import "time"

type timeSerieNode struct {
	time     time.Time
	data     interface{}
	previous *timeSerieNode
}

// StoreStatus function : stores the payload into the timeserie

// GetStatus function : calculate the diff between the last and the current payload
