package impl

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/davidwalter0/forwarder/listener"
	"github.com/davidwalter0/forwarder/mgr"
	"github.com/davidwalter0/forwarder/pipe"
	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
	empty "github.com/golang/protobuf/ptypes/empty"
	ts "github.com/golang/protobuf/ptypes/timestamp"
)

// ManagedListener control service listening socket + active connections
type ManagedListener listener.ManagedListener

// NewWatcherServer grpc code
func NewWatcherServer(Mgr *mgr.Mgr) *WatcherServer {
	return &WatcherServer{Mgr: Mgr}
}

// WatcherServer interface
type WatcherServer struct {
	Mgr *mgr.Mgr
}

var pipeGen = make(chan *pb.PipeLog)

// Now Timestamp set to current time
func Now() *ts.Timestamp {
	now := time.Now()
	s := now.Unix()
	n := int32(now.Nanosecond())
	return &ts.Timestamp{Seconds: s, Nanos: n}
}

// Generator create
func Generator() {
	for {
		pipeGen <- MockPipeGen("Gen")
		time.Sleep(share.TickDelay)
	}
}

// MockWatch could be used to test watch flow vestigial used as preliminary test code
func (w *WatcherServer) MockWatch(ignore *empty.Empty, stream pb.Watcher_WatchServer) error {
	ticker := time.NewTicker(share.TickDelay)
	defer ticker.Stop()
	go Generator()
	for {
		select {
		case pipe := <-pipeGen:
			if err := stream.Send(pipe); err != nil {
				log.Println("Watch:", err, err.Error(), reflect.TypeOf(err))
			}
		}
	}
}

// ConvertAndSend a pipe definition
func ConvertAndSend(p *pipe.Definition, stream pb.Watcher_WatchServer) {
	defer trace.Tracer.ScopedTrace()()
	if err := stream.Send(pb.Pipe2PipeLog(p)); err != nil {
		log.Println("Watch: Pipe", err, err.Error(), reflect.TypeOf(err))
	}
}

// GetPipe return the pipe info as defined
func (w *WatcherServer) GetPipe(ctx context.Context, pipeName *pb.PipeName) (*pb.PipeInfo, error) {

	// return MockPipeInfo(), nil
	defer w.Mgr.Monitor()()
	if v, ok := w.Mgr.Listeners[pipeName.Name]; ok {
		fmt.Println("v", v)
		fmt.Println("v.Definition", v.Definition)
		return pb.Pipe2PipeInfo(&v.Definition), nil
	}
	return MockPipeInfo(), nil
}

// Watch status manager to enable distributed observation via rpc
func (w *WatcherServer) Watch(ignore *empty.Empty, stream pb.Watcher_WatchServer) error {
	ticker := time.NewTicker(share.TickDelay)
	defer ticker.Stop()
	// go Generator()
	for {
		select {
		case obj, ok := <-share.Queue.Chan():
			if ok {
				switch o := obj.(type) {
				case *pipe.Definition:
					ConvertAndSend(o, stream)
				case pipe.Definition:
					ConvertAndSend(&o, stream)
				}
			} else {
				log.Println("Unable to read channel")
			}
		}
	}
}
