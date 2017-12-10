package pipe

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/go-mutex"
)

// NewPipe creates a Pipe and returns a pointer to the same
func NewPipe(name string, mapAdd, mapRm chan *Pipe, mutex *mutex.Mutex, source, sink net.Conn, definition Definition) (pipe *Pipe) {
	defer trace.Tracer.ScopedTrace()()
	defer mutex.Monitor()()
	pipe = &Pipe{Name: name,
		SourceConn: source,
		SinkConn:   sink,
		MapRm:      mapRm,
		Mutex:      mutex,
		Definition: definition,
	}
	pipe.Endpoints = &EP{}
	mapAdd <- pipe
	return
}

// Pipe a connection initiated by the return from listen and the
// up/down stream host:port pairs
type Pipe struct {
	Name       string
	SourceConn net.Conn
	SinkConn   net.Conn
	MapRm      chan *Pipe
	State      uint64
	Mutex      *mutex.Mutex
	Mode       string
	Definition
}

// Monitor lock link into
func (pipe *Pipe) Monitor(args ...interface{}) func() {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace(args...)()
		return pipe.Mutex.MonitorTrace(args...)
	}
	return func() {}
}

// Connect opens a link between source and sink
func (pipe *Pipe) Connect() {
	if pipe != nil {
		done := make(chan bool, 2)
		defer trace.Tracer.ScopedTrace()()
		go func() {
			defer trace.Tracer.ScopedTrace()()
			defer pipe.Close()
			if _, err := io.Copy(pipe.SinkConn, pipe.SourceConn); err != nil {
				log.Println(err)
			}
			done <- true
		}()
		go func() {
			defer trace.Tracer.ScopedTrace()()
			defer pipe.Close()
			if _, err := io.Copy(pipe.SourceConn, pipe.SinkConn); err != nil {
				log.Println(err)
			}
			done <- true
		}()
		go func() {
			log.Println("share.Queue.Push(&pipe.Definition)", pipe.Definition)
			share.Queue.Push(&pipe.Definition)
			for {
				ticker := time.NewTicker(share.TickDelay)
				defer ticker.Stop()
				select {
				case <-done:
					return
				case <-ticker.C:
					log.Println("share.Queue.Push(&pipe.Definition)", pipe.Definition)
					share.Queue.Push(&pipe.Definition)
				}
			}
		}()
	}
}

// Close a link between source and sink
func (pipe *Pipe) Close() {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		defer pipe.Monitor()()
		if pipe.State == share.Open {
			pipe.MapRm <- pipe
			pipe.State = share.Closed
			if err := pipe.SinkConn.Close(); err != nil {
				log.Println(err)
			}
			if err := pipe.SourceConn.Close(); err != nil {
				log.Println(err)
			}
		}
	}
}
