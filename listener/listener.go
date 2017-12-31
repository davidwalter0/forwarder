package listener

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/davidwalter0/forwarder/kubeconfig"
	"github.com/davidwalter0/forwarder/pipe"
	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-mutex"
)

var retries = 3

// ManagedListener and it's dependent objects
type ManagedListener struct {
	pipe.Definition
	Listener   net.Listener        `json:"-"`
	Pipes      map[*pipe.Pipe]bool `json:"-"`
	Mutex      *mutex.Mutex        `json:"-"`
	Wg         sync.WaitGroup      `json:"-"`
	Kubernetes bool                `json:"-"`
	Debug      bool                `json:"-"`
	n          uint64
	MapAdd     chan *pipe.Pipe
	MapRm      chan *pipe.Pipe
	StopWatch  chan bool
	Clientset  *kubernetes.Clientset
	Active     uint64
	Interface  interface{}
	InCluster  bool
}

// NewManagedListener create and populate a ManagedListener
func NewManagedListener(
	pipedef *pipe.Definition,
	envCfg *share.ServerCfg,
	Clientset *kubernetes.Clientset) (ml *ManagedListener) {
	if pipedef != nil {
		defer trace.Tracer.ScopedTrace()()
		ml = &ManagedListener{
			Definition: *pipedef,
			Listener:   Listen(pipedef.Source),
			Pipes:      make(map[*pipe.Pipe]bool),
			Mutex:      &mutex.Mutex{},
			Kubernetes: envCfg.Kubernetes,
			MapAdd:     make(chan *pipe.Pipe, 3),
			MapRm:      make(chan *pipe.Pipe, 3),
			StopWatch:  make(chan bool, 3),
			Debug:      pipedef.Debug || envCfg.Debug,
			Clientset:  Clientset,
			Active:     0,
			InCluster:  kubeconfig.InCluster,
		}
	}

	trace.Tracer.Enable(envCfg.Debug)
	return
}

// Monitor for this ManagedListener
func (ml *ManagedListener) Monitor(args ...interface{}) func() {
	if ml != nil {
		defer trace.Tracer.ScopedTrace(args...)()
		return ml.Mutex.MonitorTrace(args...)
	}
	return func() {}
}

// Insert pipe to map of pipes in managed listener
func (ml *ManagedListener) Insert(pipe *pipe.Pipe) {
	defer trace.Tracer.ScopedTrace("MapAdd", *pipe)()
	pipe.State = share.Open
	defer pipe.Monitor()()
	ml.Pipes[pipe] = true
	ml.Active = uint64(len(ml.Pipes))
}

// Delete pipe from map of pipes in managed listener
func (ml *ManagedListener) Delete(pipe *pipe.Pipe) {
	defer trace.Tracer.ScopedTrace("MapRm", *pipe)()
	pipe.State = share.Closed
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

// LoadEndpoints queries the service name for endpoints
func (ml *ManagedListener) LoadEndpoints() {
	if ml != nil && ml.InCluster {
		defer ml.Monitor()()
		var ep = pipe.EP{}
		if ep = kubeconfig.Endpoints(ml.Clientset, ml.Name, ml.Namespace); !ep.Equal(&ml.Endpoints) {
			ml.Endpoints = ep
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
		if ml.Kubernetes && ml.EnableEp && len(ml.Endpoints) > 0 {
			n = atomic.AddUint64(&ml.n, 1) % uint64(len(ml.Endpoints))
			sink = (ml.Endpoints)[n]
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
		ticker := time.NewTicker(share.TickDelay)
		defer ticker.Stop()
		for {
			select {
			case <-ml.StopWatch:
				return
			case <-ticker.C:
				ml.LoadEndpoints()
				if ml.Debug {
					log.Println(ml.Key, ml.Source, ml.Sink, ml.Name, ml.Namespace, ml.Debug, ml.Endpoints, "active", ml.Active)
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
		var pipe = pipe.NewPipe(ml.Key, ml.MapAdd, ml.MapRm, ml.Mutex, SourceConn, SinkConn, &ml.Definition)
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
			var pipes = []*pipe.Pipe{}
			for pipe := range ml.Pipes {
				pipe.Close()
				pipes = append(pipes, pipe)
			}
			for _, pipe := range pipes {
				delete(ml.Pipes, pipe)
			}
		}
	}
}

// Listen open listener on address
func Listen(address string) (listener net.Listener) {
	defer trace.Tracer.ScopedTrace()()
	var err error
	if true {
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
