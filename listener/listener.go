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

// Listen open listener on address
func Listen(address string) (listener net.Listener) {
	var err error
	// defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("listener:%v err: %v", listener, err))()
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
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
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
	Mutex      mutex.Mutex    `json:"-"`
	Wg         sync.WaitGroup `json:"-"`
	Kubernetes bool           `json:"-"`
	n          uint64
}

// NewManagedListener create and populate a ManagedListener
func NewManagedListener(pipe *PipeDefinition, kubeConfig kubeconfig.KubeConfig) *ManagedListener {
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
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
		Mutex:      mutex.Mutex{},
		Kubernetes: kubeConfig.Kubernetes,
	}
}

// Monitor for this ManagedListener
func (ml *ManagedListener) Monitor() func() {
	// defer trace.Tracer.Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("ManagedListener: %v", *ml))()
	return ml.Mutex.Monitor()
}

// Pipe a connection initiated by the return from listen and the
// up/down stream host:port pairs
type Pipe struct {
	SourceConn net.Conn
	SinkConn   net.Conn
	Pipes      *map[*Pipe]bool
	Closed     bool
}

// Open a link between source and sink
func (p *Pipe) Connect() {
	defer trace.Tracer.Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("Pipe: %v", *p))()
	go func() {
		defer trace.Tracer.Enable(trace.Enabled).ScopedTrace()()
		var err error
		defer p.Close()
		if _, err = io.Copy(p.SinkConn, p.SourceConn); err != nil {
			log.Printf("Connection failed: %v\n", err)
		}
	}()
	go func() {
		defer trace.Tracer.Enable(trace.Enabled).ScopedTrace()()
		var err error
		defer p.Close()
		if _, err = io.Copy(p.SourceConn, p.SinkConn); err != nil {
			log.Printf("Connection failed: %v\n", err)
		}
	}()
}

// Close a link between source and sink
func (p *Pipe) Close() {
	defer trace.Tracer.Enable(trace.Enabled).ScopedTrace()()
	p.SourceConn.Close()
	p.SinkConn.Close()
	pipe := *p
	delete(*pipe.Pipes, p)
}

// Open listener for this endPtDef
func (ml *ManagedListener) Open() {
	defer trace.Tracer.Enable(trace.Enabled).ScopedTrace()()
	go ml.Listening()
}

// NextEndPoint returns the next host:port pair if more than one available
// round robin selection
func (ml *ManagedListener) NextEndPoint() (sink string) {
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
	// defer trace.Tracer.Enable(trace.Enabled).ScopedTrace()()
	return ml.Listener.Accept()
}

// Listening on managed listener
func (ml *ManagedListener) Listening() {
	// defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("listener:\n%v\n", kubeconfig.Yamlify(ml.PipeDefinition)))()
	// log.Println(kubeconfig.Yamlify(ml.PipeDefinition))
	for {
		var err error
		var SourceConn, SinkConn net.Conn
		// defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace(fmt.Sprintf("listener:%v", ml))()
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
		pipe := &Pipe{SourceConn: SourceConn, SinkConn: SinkConn, Pipes: &ml.Pipes}
		ml.Pipes[pipe] = true
		go pipe.Connect()
	}
}

// Close a listener and it's children
func (ml *ManagedListener) Close() {
	if err := ml.Listener.Close(); err != nil {
		log.Println("Error closing listener", ml.Listener)
	}
	defer ml.Monitor()()
	for pipe := range ml.Pipes {
		pipe.Close()
	}
}
