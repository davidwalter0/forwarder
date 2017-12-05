package chanqueue

import (
	"fmt"
)

// ChanQueue with non-blocking push/pop interface
type ChanQueue struct {
	Channel chan interface{}
	isOpen  bool
}

// NewChanQueue general channel with push / pop of interface
func NewChanQueue() *ChanQueue {
	return &ChanQueue{Channel: make(chan interface{}, 8), isOpen: true}
}

// Close the channel
func (cq *ChanQueue) Close() (err error) {
	if cq.IsOpen() {
		close(cq.Channel)
		cq.isOpen = false
	} else {
		err = fmt.Errorf("Pipeline channel already closed")
	}
	return
}

// IsOpen status of channel
func (cq *ChanQueue) IsOpen() bool {
	return cq.isOpen
}

// Push data from pipe info
func (cq *ChanQueue) Push(m interface{}) {
	if cq.IsOpen() {
		select {
		case cq.Channel <- m:
		default:
		}
	}
}

// Pop data from pipe info
func (cq *ChanQueue) Pop() (m interface{}, ok bool) {
	if cq.IsOpen() {
		select {
		case m, ok = <-cq.Channel:
			return m, ok
		default:
		}
	}
	return
}

// Chan return the underlying blocking channel
func (cq *ChanQueue) Chan() chan interface{} {
	return cq.Channel
}
