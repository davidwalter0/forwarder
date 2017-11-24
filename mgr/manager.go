// New work flow
// RUN
// - load/reload kubeConfig
// - if not listening, create new listener

//   - new connection
//     - add connection pair to pipe list

// Allow existing connections to persist until closed even when the
// kubeConfig is been removed - defer removal code
//     - run go routine with args pipe & remove method
//     - on close remove pipe record from mgr
//   - close changed listener's connections
//     - create go routine to close new items
//   - run cleanup close loop
package mgr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/davidwalter0/forwarder/kubeconfig"
	"github.com/davidwalter0/forwarder/listener"
	"github.com/davidwalter0/forwarder/set"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-cfg"
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

var kubeConfig kubeconfig.KubeConfig

var reload = make(chan os.FileInfo)
var delta = time.Duration(5)

// Version semver string
var Version string // = strings.Split(string(Load(".version")), "=")[1]

// ManagedListener control service listening socket + active connections
type ManagedListener listener.ManagedListener

// PipeDefinition for service source / sink
// type PipeDefinitions map[string]*listener.PipeDefinition

// NewManagedListener object create
var NewManagedListener = listener.NewManagedListener

var pipeDefs *map[string]*listener.PipeDefinition

// Mgr management info for listeners
type Mgr struct {
	Listeners map[string]*listener.ManagedListener
	Mutex     mutex.Mutex
}

// Monitor lifts mutex deferable lock to Mgr object
func (mgr *Mgr) Monitor() func() {
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
	return mgr.Mutex.Monitor()
}

// Monitor lifts mutex deferable lock to Mgr object
func (mgr *Mgr) LoadEndpoints() {
	for k, v := range mgr.Listeners {
		if v != nil {
			if v.EnableEp {
				v.Endpoints = kubeconfig.Endpoints(v.Service, v.Namespace)
				if kubeConfig.Debug {
					log.Println("mgr", k, v.Service, v.Namespace, v.Endpoints, v.Source, v.Sink)
				}
			}
		}
	}
}

// Run primary processing loop
func (mgr *Mgr) Run() {
	Configure()
	mgr.Listeners = make(map[string]*listener.ManagedListener)
	// defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
	{
		var m = make(map[string]*listener.PipeDefinition)
		mgr.Merge(&m)
	}
	go Watch()
	var pipeDefs = LoadEndPts()
	for {
		{
			if kubeConfig.Debug {
				log.Println("Loop in Run()")
			}
			select {
			case stat := <-reload:
				// defer trace.Tracer.Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("Reload %v %v", stat.Name(), stat.ModTime()))()
				if kubeConfig.Debug {
					log.Printf("Reload %v %v\n", stat.Name(), stat.ModTime())
				}
				mgr.Merge(pipeDefs)
			case delay := <-time.After(time.Second * logReloadTimeout):
				if kubeConfig.Debug {
					log.Printf("Reload timed out after %d seconds %v\n", logReloadTimeout, delay)
				}
			}
		}
	}
}

// Merge the kubeConfiguration of pipeDefs
func (mgr *Mgr) Merge(lhs *map[string]*listener.PipeDefinition) {
	defer mgr.Monitor()()
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
	rhs := LoadEndPts()
	var LOnly, Common, ROnly = set.Difference(lhs, rhs)
	// Not Common or in the right hand (new kubeConfig) set, are now
	// vestiges of the prior (lhs) set
	for _, k := range LOnly {
		{
			log.Println("closing lhs[k]", k, (*lhs)[k])
			mgr.Listeners[k].Close()
			delete((*lhs), k)
			delete(mgr.Listeners, k)
		}
	}

	// If Common names were updated, replace with new kubeConfig
	for _, k := range Common {
		{
			log.Println("common lhs[k]", k, (*lhs)[k], "equal", !(*lhs)[k].Equal((*rhs)[k]))
			if !(*lhs)[k].Equal((*rhs)[k]) {
				mgr.Listeners[k].Close()
				delete((*lhs), k)
				delete(mgr.Listeners, k)
				mgr.Listeners[k] = NewManagedListener((*rhs)[k], kubeConfig)
				(*lhs)[k] = listener.NewPipeDefinition((*rhs)[k])
				mgr.Listeners[k].Open()
			}
		}
	}

	// Add new items (not in L, existing)
	for _, k := range ROnly {
		log.Println("right only rhs[k]", k, (*rhs)[k])
		(*lhs)[k] = listener.NewPipeDefinition((*rhs)[k])
		mgr.Listeners[k] = NewManagedListener((*rhs)[k], kubeConfig)
		mgr.Listeners[k].Open()
	}
	mgr.LoadEndpoints()
}

var complete = make(chan bool)
var counter uint64

// CheckInCluster reports if the env variable is set for cluster
func CheckInCluster() bool {
	return len(os.Getenv("KUBERNETES_PORT")) > 0
}

// Configure from environment variables or command line flags and load
// the kubeConfiguration file endPtDefing pairs.
func Configure() {
	var err error

	if err = cfg.Process("", &kubeConfig); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(kubeConfig.File) == 0 {
		fmt.Println("Error: kubeConfiguration file not set")
		cfg.Usage()
		os.Exit(1)
	}

	var jsonText []byte
	jsonText, _ = json.MarshalIndent(&kubeConfig, "", "  ")
	if kubeConfig.Debug {
		fmt.Printf("\n%v\n", string(jsonText))
	}
	kubeConfig.LoadKubeConfig()
	trace.Enabled = kubeConfig.Debug
}

// LoadEndPts load from text into a pipeDefs object
func LoadEndPts() (e *map[string]*listener.PipeDefinition) {
	var m = make(map[string]*listener.PipeDefinition)
	e = &m
	var text = Load(kubeConfig.File)
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
func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			// Secret update -> REMOVE event, invalidates the watch,
			// reassert
			err = watcher.Add(kubeConfig.File)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 3)
			} else {
				select {
				case event := <-watcher.Events:
					stat, err := os.Stat(kubeConfig.File)
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
					stat, e := os.Stat(kubeConfig.File)
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
