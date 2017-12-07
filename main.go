//go:generate mockgen -destination mock/mock_forwarder.go github.com/davidwalter0/listener github.com/davidwalter0/forwarder Insert,Listen,Close,Listening,EpWatcher,StopWatchNotify,LoadEndpoints
// New work flow
// RUN
// - load/reload envCfg
// - if not listening, create new listener

//   - new connection
//     - add connection pair to pipe list

// Allow existing connections to persist until closed even when the
// envCfg is been removed - defer removal code
//     - run go routine with args pipe & remove method
//     - on close remove pipe record from mgr
//   - close changed listener's connections
//     - create go routine to close new items
//   - run cleanup close loop
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davidwalter0/forwarder/listener"
	mgmt "github.com/davidwalter0/forwarder/mgr"
	"github.com/davidwalter0/forwarder/share"
)

var complete = make(chan bool)
var envCfg = &share.ServerCfg{}
var mgr *mgmt.Mgr = mgmt.NewMgr(envCfg)

// retries number of attempts
var retries = 3

// logReloadTimeout in seconds
var logReloadTimeout = time.Duration(600)

// Build info text
var Build string

// Commit git string
var Commit string

// Version semver string
var Version string

// ManagedListener control service listening socket + active connections
type ManagedListener listener.ManagedListener

func main() {
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Printf("%s: Version %s version build %s commit %s\n", me, Version, Build, Commit)
	envCfg.Read()
	mgr.Run()
	<-complete
}
