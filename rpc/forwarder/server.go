//go:generate protoc --proto_path=pipe --proto_path=/go/src --go_out=plugins=grpc:pipe pipe/pipe.proto

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	mgmt "github.com/davidwalter0/forwarder/mgr"
	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	impl "github.com/davidwalter0/forwarder/rpc/server/impl"
	"github.com/davidwalter0/forwarder/share"
	// log "github.com/davidwalter0/logwriter"
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

var complete = make(chan bool)
var envCfg = &share.ServerCfg{}
var mgr = mgmt.NewMgr(envCfg)

// retries number of attempts
var retries = 3

// logReloadTimeout in seconds
var logReloadTimeout = time.Duration(share.TickDelay)

func main() {
	envCfg.Read()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s", envCfg.ServerAddr))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if envCfg.TLS {
		creds := envCfg.LoadServerCreds()
		opts = append(opts, creds)
	}

	go func() {
		grpcServer := grpc.NewServer(opts...)
		pb.RegisterWatcherServer(grpcServer, impl.NewWatcherServer(mgr))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	mgr.Run()
	<-complete
}
