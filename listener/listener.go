package listener

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/davidwalter0/forwarder/kubeconfig"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-mutex"
)

var retries = 3

func init() {
	// trace.Tracer.Detailed(true).Enable(true)
}

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
	Source    string   `json:"source"    help:"source ingress point host:port"`
	Sink      string   `json:"sink"      help:"sink service point   host:port"`
	Endpoints []string `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
	EnableEp  bool     `json:"enable-ep" help:"enable endpoints from service"`
	Service   string   `json:"service"   help:"service name"`
	Namespace string   `json:"namespace" help:"service namespace"`
}

// NewPipeDefinition create and initialize a PipeDefinition
func NewPipeDefinition(pipe *PipeDefinition) *PipeDefinition {
	defer trace.Tracer.ScopedTrace()()
	return &PipeDefinition{
		Source:    pipe.Source,
		Sink:      pipe.Sink,
		EnableEp:  pipe.EnableEp,
		Service:   pipe.Service,
		Namespace: pipe.Namespace,
	}
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
	n          uint64
	MapAdd     chan *Pipe
	MapRm      chan *Pipe
	MapClear   chan bool
}

// NewManagedListener create and populate a ManagedListener
func NewManagedListener(pipe *PipeDefinition, kubeConfig kubeconfig.KubeConfig) *ManagedListener {
	defer trace.Tracer.ScopedTrace()()
	return &ManagedListener{
		PipeDefinition: *pipe,
		// PipeDefinition: PipeDefinition{
		// 	Source:    pipe.Source,
		// 	Sink:      pipe.Sink,
		// 	EnableEp:  pipe.EnableEp,
		// 	Service:   pipe.Service,
		// 	Namespace: pipe.Namespace,
		// },
		Listener:   Listen(pipe.Source),
		Pipes:      make(map[*Pipe]bool),
		Mutex:      &mutex.Mutex{},
		Kubernetes: kubeConfig.Kubernetes,
		MapAdd:     make(chan *Pipe, 3),
		MapRm:      make(chan *Pipe, 3),
		MapClear:   make(chan bool),
	}
}

// NewPipe creates a Pipe and returns a pointer to the same
func NewPipe(ml *ManagedListener, source, sink net.Conn) (pipe *Pipe) {
	defer ml.Monitor()()
	pipe = &Pipe{SourceConn: source, SinkConn: sink, MapRm: ml.MapRm, Mutex: ml.Mutex}
	ml.MapAdd <- pipe
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
	SourceConn net.Conn
	SinkConn   net.Conn
	MapRm      chan *Pipe
	State      uint64
	Mutex      *mutex.Mutex
}

// Monitor lock link into
func (pipe *Pipe) Monitor(args ...interface{}) func() {
	defer trace.Tracer.ScopedTrace(args...)()
	return pipe.Mutex.MonitorTrace(args...)
}

// Monitor for this ManagedListener
func (ml *ManagedListener) Monitor(args ...interface{}) func() {
	defer trace.Tracer.ScopedTrace(args...)()
	// defer trace.Tracer.ScopedTrace(fmt.Sprintf("ManagedListener: %v", *ml))()
	return ml.Mutex.MonitorTrace(args...)
}

// Connect opens a link between source and sink
func (pipe *Pipe) Connect() {
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

// Close a link between source and sink
func (pipe *Pipe) Close() {
	defer trace.Tracer.ScopedTrace()()
	exit := pipe.Monitor()
	if pipe.State == Open {
		pipe.SinkConn.Close()
		pipe.SourceConn.Close()
		pipe.MapRm <- pipe
	}
	exit()
}

// PipeMapHandler adds, removes, closes and single threads access to map list
func (ml *ManagedListener) PipeMapHandler() {
	for {
		select {
		case pipe := <-ml.MapAdd:
			{
				exit := trace.Tracer.ScopedTrace("MapAdd", *pipe)
				pipe.State = Open
				ml.Pipes[pipe] = true
				exit()
			}
		case pipe := <-ml.MapRm:
			{
				exit := trace.Tracer.ScopedTrace("MapRm", *pipe)
				pipe.State = Closed
				delete(ml.Pipes, pipe)
				exit()
			}
		}
	}
}

// Open listener for this endPtDef
func (ml *ManagedListener) Open() {
	defer trace.Tracer.ScopedTrace()()
	go ml.Listening()
	go ml.PipeMapHandler()
}

// NextEndPoint returns the next host:port pair if more than one available
// round robin selection
func (ml *ManagedListener) NextEndPoint() (sink string) {
	defer trace.Tracer.ScopedTrace()()
	var n uint64
	// Don't use k8s endpoint lookup if not in a k8s cluster
	if ml.Kubernetes && ml.EnableEp && len(ml.Endpoints) > 0 {
		n = atomic.AddUint64(&ml.n, 1) % uint64(len(ml.Endpoints))
		sink = ml.Endpoints[n]
	} else {
		sink = ml.Sink
	}
	return
}

// Accept expose ManagedListener's listener
func (ml *ManagedListener) Accept() (net.Conn, error) {
	defer trace.Tracer.ScopedTrace()()
	return ml.Listener.Accept()
}

// Listening on managed listener
func (ml *ManagedListener) Listening() {
	defer trace.Tracer.ScopedTrace()()
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
		var pipe = NewPipe(ml, SourceConn, SinkConn)
		go pipe.Connect()
	}
}

// Close a listener and it's children
func (ml *ManagedListener) Close() {
	defer trace.Tracer.ScopedTrace()()
	if err := ml.Listener.Close(); err != nil {
		log.Println("Error closing listener", ml.Listener)
	}
	defer ml.Monitor()()
	for pipe := range ml.Pipes {
		pipe.Close()
	}
}
