// Package mgr manages listeners for each forwarding pipe definition
package mgr

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/davidwalter0/forwarder/listener"
	"github.com/davidwalter0/forwarder/pipe"
	"github.com/davidwalter0/forwarder/set"
	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-mutex"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// retries number of attempts
var retries = 3

// logReloadTimeout in seconds
var logReloadTimeout = time.Duration(600)

// Build info text
var Build string

// Commit git string
var Commit string

var reload = make(chan os.FileInfo)
var delta = time.Duration(5)

// Version semver string
var Version string // = strings.Split(string(Load(".version")), "=")[1]

// ManagedListener control service listening socket + active connections
type ManagedListener listener.ManagedListener

// PipeDefinition for service source / sink
// type PipeDefinitions map[string]*pipe.Definition

// NewManagedListener object create
var NewManagedListener = listener.NewManagedListener

// NewFromDefinition object create
var NewFromDefinition = pipe.NewFromDefinition

var pipeDefs *map[string]*pipe.Definition

// Mgr management info for listeners
type Mgr struct {
	Listeners map[string]*listener.ManagedListener
	Mutex     *mutex.Mutex
	EnvCfg    *share.ServerCfg
}

// NewMgr create a new Mgr
func NewMgr(EnvCfg *share.ServerCfg) *Mgr {
	return &Mgr{
		Listeners: make(map[string]*listener.ManagedListener),
		Mutex:     &mutex.Mutex{},
		EnvCfg:    EnvCfg,
	}
}

// Monitor lifts mutex deferable lock to Mgr object
func (mgr *Mgr) Monitor() func() {
	defer trace.Tracer.ScopedTrace()()
	return mgr.Mutex.MonitorTrace()
}

// Run primary processing loop
func (mgr *Mgr) Run() {
	var EnvCfg = mgr.EnvCfg
	{
		var m = make(map[string]*pipe.Definition)
		mgr.Merge(&m)
	}
	go mgr.Watch()
	var pipeDefs = mgr.LoadEndPts()
	for {
		{
			if EnvCfg.Debug {
				log.Println("Loop in Run()")
			}
			select {
			case stat := <-reload:
				if EnvCfg.Debug {
					log.Printf("Reload %v %v\n", stat.Name(), stat.ModTime())
				}
				mgr.Merge(pipeDefs)
			case delay := <-time.After(time.Second * logReloadTimeout):
				if EnvCfg.Debug {
					log.Printf("Reload timed out after %d seconds %v\n", logReloadTimeout, delay)
				}
			}
		}
	}
}

// Merge the EnvCfguration of pipeDefs
func (mgr *Mgr) Merge(lhs *map[string]*pipe.Definition) {
	defer mgr.Mutex.MonitorTrace("Merge")()
	defer trace.Tracer.ScopedTrace()()
	var EnvCfg = mgr.EnvCfg
	rhs := mgr.LoadEndPts()
	var LOnly, Common, ROnly = set.Difference(lhs, rhs)
	// Not Common or in the right hand (new EnvCfg) set, are now
	// vestiges of the prior (lhs) set
	for _, k := range LOnly {
		log.Println("closing lhs[k]", k, (*lhs)[k])
		mgr.Listeners[k].Close()
		delete((*lhs), k)
		delete(mgr.Listeners, k)
	}

	// If Common names were updated, replace with new EnvCfg
	for _, k := range Common {
		log.Println("common lhs[k]", k, (*lhs)[k], "equal", !(*lhs)[k].Equal((*rhs)[k]))
		if !(*lhs)[k].Equal((*rhs)[k]) {
			if listener, ok := mgr.Listeners[k]; ok && listener != nil {
				mgr.Listeners[k].Close()
				delete((*lhs), k)
				delete(mgr.Listeners, k)
				(*rhs)[k].Name = k
				(*lhs)[k] = NewFromDefinition((*rhs)[k])
				(*lhs)[k].Name = k
				mgr.Listeners[k] = NewManagedListener((*lhs)[k], EnvCfg)
				mgr.Listeners[k].Open()
			}
		}
	}

	// Add new items (not in L, existing)
	for _, k := range ROnly {
		log.Println("right only rhs[k]", k, (*rhs)[k])
		(*rhs)[k].Name = k
		(*lhs)[k] = NewFromDefinition((*rhs)[k])
		(*lhs)[k].Name = k
		mgr.Listeners[k] = NewManagedListener((*lhs)[k], EnvCfg)
		mgr.Listeners[k].Open()
	}
	// mgr.LoadEndpoints()
}

var complete = make(chan bool)
var counter uint64

// CheckInCluster reports if the env variable is set for cluster
func CheckInCluster() bool {
	return len(os.Getenv("KUBERNETES_PORT")) > 0
}

// LoadEndPts load from text into a pipeDefs object
func (mgr *Mgr) LoadEndPts() (e *map[string]*pipe.Definition) {
	var m = make(map[string]*pipe.Definition)
	e = &m
	var text = Load(mgr.EnvCfg.File)
	var err = yaml.Unmarshal(text, e)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return
}

// Load helper function from file to []byte
func Load(filename string) []byte {
	if len(filename) == 0 {
		panic(fmt.Sprintf("Can't Load() a file with an empty name"))
	}

	var err error
	var text []byte
	if len(filename) > 0 {
		text, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("%v\n", err)
			os.Exit(3)
		}
	}
	return text
}

// Watch reports file change
func (mgr *Mgr) Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			// Secret update -> REMOVE event, invalidates the watch,
			// reacquire by recreating
			err = watcher.Add(mgr.EnvCfg.File)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 3)
			} else {
				select {
				case event := <-watcher.Events:
					stat, err := os.Stat(mgr.EnvCfg.File)
					reload <- stat
					msg := fmt.Sprintf("Reload %v %v", stat.Name(), stat.ModTime())
					if err == nil {
						log.Printf("stat %v %v %v\n", event.Name, msg, err)
					} else {
						log.Printf("stat %v %v %v\n", event.Name, msg, err)
					}
					op := event.Op
					if op&fsnotify.Create == fsnotify.Create {
						log.Println("CREATE", event.Name, msg)
					}
					if op&fsnotify.Remove == fsnotify.Remove {
						log.Println("REMOVE", event.Name, msg)
					}
					if op&fsnotify.Write == fsnotify.Write {
						log.Println("WRITE", event.Name, msg)
					}
					if op&fsnotify.Rename == fsnotify.Rename {
						log.Println("RENAME", event.Name, msg)
					}
					if op&fsnotify.Chmod == fsnotify.Chmod {
						log.Println("CHMOD", event.Name, msg)
					}
				case err := <-watcher.Errors:
					stat, e := os.Stat(mgr.EnvCfg.File)
					msg := fmt.Sprintf("watch error %v %v %v %v", err, stat.Name(), stat.ModTime(), e)
					log.Println("error:", msg)
				}
			}
		}
	}()

	defer watcher.Close()
	done := make(chan bool)
	<-done
}
