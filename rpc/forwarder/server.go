//go:generate protoc --proto_path=pipe --proto_path=/go/src --go_out=plugins=grpc:pipe pipe/pipe.proto

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"

	"crypto/tls"
	"crypto/x509"

	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	empty "github.com/golang/protobuf/ptypes/empty"
	ts "github.com/golang/protobuf/ptypes/timestamp"

	"github.com/davidwalter0/forwarder/listener"
	mgmt "github.com/davidwalter0/forwarder/mgr"
	"github.com/davidwalter0/forwarder/pipe"
	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
)

var (
	withTLS    = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	rootCAFile = flag.String("root-ca", "certs/RootCA.crt", "The Root CA file")
	certFile   = flag.String("cert-file", "certs/example.com.crt", "The TLS cert file")
	keyFile    = flag.String("key-file", "certs/example.com.key", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "testdata/route_guide_db.json", "A json file containing a list of features")
	serverAddr = flag.String("server-addr", "0.0.0.0:10000", "The server address in the format of host:port")
)

// Build info text
var Build string

// Commit git string
var Commit string

// Version semver string
var Version string

func init() {
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Printf("%s: Version %s version build %s commit %s\n", me, Version, Build, Commit)
}

// NewWatcherServer grpc code
func NewWatcherServer() *WatcherServer {
	return new(WatcherServer)
}

func loadCreds() grpc.ServerOption {

	certificate, err := tls.LoadX509KeyPair(
		*certFile,
		*keyFile,
	)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(*rootCAFile)
	if err != nil {
		log.Fatalf("failed to read client ca cert: %s", err)
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatal("failed to append client certs")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}

	serverOption := grpc.Creds(credentials.NewTLS(tlsConfig))
	return serverOption
}

// WatcherServer interface
type WatcherServer struct {
}

var pipeGen = make(chan *pb.PipeLog)

// Now Timestamp set to current time
func Now() *ts.Timestamp {
	now := time.Now()
	s := now.Unix()
	n := int32(now.Nanosecond())
	return &ts.Timestamp{Seconds: s, Nanos: n}
}

// MockPipeGen create an example pipe
func MockPipeGen(which string) *pb.PipeLog {
	return &pb.PipeLog{
		Timestamp: Now(),
		Text:      which,
		PipeInfo: &pb.PipeInfo{
			Name:      "Example",
			Source:    "0.0.0.0:80",
			Sink:      "10.30.0.80",
			EnableEp:  false,
			Service:   "",
			Namespace: "",
			Debug:     false,
			Endpoints: []string{},
			Mode:      pb.Mode_P2P,
		},
	}
}

/*
// Generator create
func Generator() {
	for {
		pipeGen <- MockPipeGen("Gen")
		time.Sleep(time.Second * 30)
	}
}

func (l *WatcherServer) Watch(ignore *empty.Empty, stream pb.Watcher_WatchServer) error {
	ticker := time.NewTicker(5 * time.Second)
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
*/
func ConvertAndSend(p *pipe.Definition, stream pb.Watcher_WatchServer) {
	defer trace.Tracer.ScopedTrace()()
	if err := stream.Send(pb.Pipe2PipeLog(p)); err != nil {
		log.Println("Watch: Pipe", err, err.Error(), reflect.TypeOf(err))
	}
}

// Watch status manager to enable distributed observation via rpc
func (l *WatcherServer) Watch(ignore *empty.Empty, stream pb.Watcher_WatchServer) error {
	ticker := time.NewTicker(5 * time.Second)
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

// ManagedListener control service listening socket + active connections
type ManagedListener listener.ManagedListener

var complete = make(chan bool)
var mgr *mgmt.Mgr = mgmt.NewMgr()

// retries number of attempts
var retries = 3

// logReloadTimeout in seconds
var logReloadTimeout = time.Duration(600)

func main() {
	// flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s", *serverAddr))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *withTLS {
		opts = append(opts, loadCreds())
	}
	go func() {
		grpcServer := grpc.NewServer(opts...)
		pb.RegisterWatcherServer(grpcServer, NewWatcherServer())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	mgr.Run()
	<-complete
}
