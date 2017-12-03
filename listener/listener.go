package listener

import (
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davidwalter0/forwarder/kubeconfig"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-mutex"
)

var retries = 3

// func init() {
// 	trace.Tracer.Detailed(true).Enable(true)
// }

// Listen open listener on address
func Listen(address string) (listener net.Listener) {
	defer trace.Tracer.ScopedTrace()()
	var err error
	if false {
		defer trace.Tracer.ScopedTrace(fmt.Sprintf("listener:%v err: %v", listener, err))()
	}
	for i := 0; i < retries; i++ {
		listener, err = net.Listen("tcp", address)
		if err != nil {
			log.Printf("net.Listen(\"tcp\", %s ) failed: %v\n", address, err)
		} else {
			return listener
		}
	}
	return
}

// PipeDefinition maps source to sink
type PipeDefinition struct {
	Name      string `json:"name"      help:"map key"`
	Source    string `json:"source"    help:"source ingress point host:port"`
	Sink      string `json:"sink"      help:"sink service point   host:port"`
	Endpoints *EP    `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
	EnableEp  bool   `json:"enable-ep" help:"enable endpoints from service"`
	Service   string `json:"service"   help:"service name"`
	Namespace string `json:"namespace" help:"service namespace"`
	Debug     bool   `json:"debug"     help:"enable debug for this pipe"`
}

// NewPipeDefinition create and initialize a PipeDefinition
func NewPipeDefinition(pipe *PipeDefinition) (pipeDefinition *PipeDefinition) {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		pipeDefinition = &PipeDefinition{
			// Name is the key of yaml map
			// Name:      name,
			Source:    pipe.Source,
			Sink:      pipe.Sink,
			EnableEp:  pipe.EnableEp,
			Service:   pipe.Service,
			Namespace: pipe.Namespace,
			Debug:     pipe.Debug,
		}
	}
	return
}

// PipeDefinitions from text description in yaml
type PipeDefinitions map[string]*PipeDefinition

// ManagedListener and it's dependent objects
type ManagedListener struct {
	PipeDefinition
	Listener   net.Listener   `json:"-"`
	Pipes      map[*Pipe]bool `json:"-"`
	Mutex      *mutex.Mutex   `json:"-"`
	Wg         sync.WaitGroup `json:"-"`
	Kubernetes bool           `json:"-"`
	Debug      bool           `json:"-"`
	n          uint64
	MapAdd     chan *Pipe
	MapRm      chan *Pipe
	StopWatch  chan bool
	Active     uint64
}

// NewManagedListener create and populate a ManagedListener
func NewManagedListener(pipe *PipeDefinition, kubeConfig kubeconfig.Cfg) (ml *ManagedListener) {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		ml = &ManagedListener{
			PipeDefinition: *pipe,
			Listener:       Listen(pipe.Source),
			Pipes:          make(map[*Pipe]bool),
			Mutex:          &mutex.Mutex{},
			Kubernetes:     kubeConfig.Kubernetes,
			MapAdd:         make(chan *Pipe, 3),
			MapRm:          make(chan *Pipe, 3),
			StopWatch:      make(chan bool, 3),
			Debug:          pipe.Debug || kubeConfig.Debug,
			Active:         0,
		}
	}
	return
}

// NewPipe creates a Pipe and returns a pointer to the same
func NewPipe(name string, ml *ManagedListener, source, sink net.Conn) (pipe *Pipe) {
	if ml != nil {
		defer ml.Monitor()()
		pipe = &Pipe{Name: name, SourceConn: source, SinkConn: sink, MapRm: ml.MapRm, Mutex: ml.Mutex}
		ml.MapAdd <- pipe
	}
	return
}

const (
	// Open : State
	Open = iota
	// Closed : State
	Closed
)

// Pipe a connection initiated by the return from listen and the
// up/down stream host:port pairs
type Pipe struct {
	Name       string
	SourceConn net.Conn
	SinkConn   net.Conn
	MapRm      chan *Pipe
	State      uint64
	Mutex      *mutex.Mutex
}

// Monitor lock link into
func (pipe *Pipe) Monitor(args ...interface{}) func() {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace(args...)()
		return pipe.Mutex.MonitorTrace(args...)
	}
	return func() {}
}

// Monitor for this ManagedListener
func (ml *ManagedListener) Monitor(args ...interface{}) func() {
	if ml != nil {
		defer trace.Tracer.ScopedTrace(args...)()
		return ml.Mutex.MonitorTrace(args...)
	}
	return func() {}
}

// Connect opens a link between source and sink
func (pipe *Pipe) Connect() {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		go func() {
			defer trace.Tracer.ScopedTrace()()
			defer pipe.Close()
			io.Copy(pipe.SinkConn, pipe.SourceConn)
		}()
		go func() {
			defer trace.Tracer.ScopedTrace()()
			defer pipe.Close()
			io.Copy(pipe.SourceConn, pipe.SinkConn)
		}()
	}
}

// Close a link between source and sink
func (pipe *Pipe) Close() {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		defer pipe.Monitor()()
		if pipe.State == Open {
			pipe.MapRm <- pipe
			pipe.State = Closed
			pipe.SinkConn.Close()
			pipe.SourceConn.Close()
		}
	}
}

// Insert pipe to map of pipes in managed listener
func (ml *ManagedListener) Insert(pipe *Pipe) {
	defer trace.Tracer.ScopedTrace("MapAdd", *pipe)()
	pipe.State = Open
	defer pipe.Monitor()()
	ml.Pipes[pipe] = true
	ml.Active = uint64(len(ml.Pipes))
}

// Delete pipe from map of pipes in managed listener
func (ml *ManagedListener) Delete(pipe *Pipe) {
	defer trace.Tracer.ScopedTrace("MapRm", *pipe)()
	pipe.State = Closed
	defer pipe.Monitor()()
	delete(ml.Pipes, pipe)
	ml.Active = uint64(len(ml.Pipes))
}

// PipeMapHandler adds, removes, closes and single threads access to map list
func (ml *ManagedListener) PipeMapHandler() {
	if ml != nil {
		for {
			select {
			case pipe := <-ml.MapAdd:
				ml.Insert(pipe)
			case pipe := <-ml.MapRm:
				ml.Delete(pipe)
			}
		}
	}
}

// Open listener for this endPtDef
func (ml *ManagedListener) Open() {
	if ml != nil {
		defer trace.Tracer.ScopedTrace()()
		go ml.Listening()
		go ml.PipeMapHandler()
	}
}

// EP slice of endpoints
type EP []string

// Equal compare two endpoint arrays for equality
func (ep *EP) Equal(rhs *EP) (rc bool) {
	if ep != nil && rhs != nil && len(*ep) == len(*rhs) {
		sort.Strings(*ep)
		sort.Strings(*rhs)
		for i, v := range *ep {
			if v != (*rhs)[i] {
				return
			}
		}
	} else {
		return
	}
	return true
}

// LoadEndpoints queries the service name for endpoints
func (ml *ManagedListener) LoadEndpoints() {
	if ml != nil {
		defer ml.Monitor()()
		var ep EP = EP{}
		if ep = kubeconfig.Endpoints(ml.Service, ml.Namespace); !ep.Equal(ml.Endpoints) {
			ml.Endpoints = &ep
		}
	}
}

// NextEndPoint returns the next host:port pair if more than one
// available round robin selection
func (ml *ManagedListener) NextEndPoint() (sink string) {
	if ml != nil {
		defer trace.Tracer.ScopedTrace()()
		defer ml.Monitor()()
		var n uint64
		// Don't use k8s endpoint lookup if not in a k8s cluster
		if ml.Kubernetes && ml.EnableEp && len(*ml.Endpoints) > 0 {
			n = atomic.AddUint64(&ml.n, 1) % uint64(len(*ml.Endpoints))
			sink = (*ml.Endpoints)[n]
		} else {
			sink = ml.Sink
		}
	}
	return
}

// Accept expose ManagedListener's listener
func (ml *ManagedListener) Accept() (net.Conn, error) {
	defer trace.Tracer.ScopedTrace()()
	return ml.Listener.Accept()
}

// StopWatchNotify checking for endpoints
func (ml *ManagedListener) StopWatchNotify() {
	if ml != nil {
		ml.StopWatch <- true
	}
}

// EpWatcher check for endpoints
func (ml *ManagedListener) EpWatcher() {
	if ml != nil {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ml.StopWatch:
				return
			case <-ticker.C:
				ml.LoadEndpoints()
				if ml.Debug {
					log.Println(ml.Name, ml.Source, ml.Sink, ml.Service, ml.Namespace, ml.Debug, *ml.Endpoints, "active", ml.Active)
				}
			}
		}
	}
}

// Listening on managed listener
func (ml *ManagedListener) Listening() {
	defer trace.Tracer.ScopedTrace()()
	defer ml.StopWatchNotify()
	go ml.EpWatcher()
	for {
		var err error
		var SourceConn, SinkConn net.Conn
		// defer trace.Tracer.ScopedTrace(fmt.Sprintf("listener:%v", ml))()
		if SourceConn, err = ml.Accept(); err != nil {
			log.Printf("Connection failed: %v\n", err)
			break
		}
		sink := ml.NextEndPoint()
		SinkConn, err = net.Dial("tcp", sink)
		if err != nil {
			log.Printf("Connection failed: %v\n", err)
			break
		}
		var pipe = NewPipe(ml.Name, ml, SourceConn, SinkConn)
		go pipe.Connect()
	}
}

// Close a listener and it's children
func (ml *ManagedListener) Close() {
	if ml != nil {
		defer trace.Tracer.ScopedTrace()()
		if ml.Listener != nil {
			if err := ml.Listener.Close(); err != nil {
				log.Println("Error closing listener", ml.Listener)
			}
			defer ml.Monitor()()
			for pipe := range ml.Pipes {
				pipe.Close()
			}
		}
	}
}
