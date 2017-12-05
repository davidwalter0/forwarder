package share

import (
	"time"

	"github.com/davidwalter0/forwarder/chanqueue"
)

const (
	// TickDelay delay between log entries
	TickDelay = time.Duration(1)
	// Open : State
	Open = iota
	// Closed : State
	Closed
)

var Queue = chanqueue.NewChanQueue()
