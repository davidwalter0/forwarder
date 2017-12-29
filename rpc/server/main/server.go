//go:generate protoc --proto_path=pipe --proto_path=/go/src --go_out=plugins=grpc:pipe pipe/pipe.proto

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	mgmt "github.com/davidwalter0/forwarder/mgr"
	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	"github.com/davidwalter0/forwarder/rpc/server/impl"
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
var envCfg = share.NewServerCfg()
var mgr = mgmt.NewMgr(envCfg)

func main() {
	envCfg.Read()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", envCfg.ServerAddr))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if envCfg.TLS {
		opts = append(opts, envCfg.LoadServerCreds())
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterWatcherServer(grpcServer, impl.NewWatcherServer(mgr))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
	mgr.Run()
	<-complete
}
