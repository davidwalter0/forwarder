package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/davidwalter0/forwarder/rpc/client/impl"
	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	"github.com/davidwalter0/forwarder/share"
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

var done = make(chan bool)

func main() {
	var clientCfg = &share.ClientCfg{}
	clientCfg.Read()

	var opts []grpc.DialOption
	if clientCfg.TLS {
		creds := clientCfg.LoadClientCreds()
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(clientCfg.ServerAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()
	client := pb.NewWatcherClient(conn)
	go func() {
		for {
			impl.RunPipeLogClient(client)
		}
	}()
	go func() {
		for {
			var pipeName *pb.PipeName
			pipeName = &pb.PipeName{Name: "echo"}
			impl.RunPipeInfoRequest(client, pipeName)
			pipeName = &pb.PipeName{Name: "ssh"}
			impl.RunPipeInfoRequest(client, pipeName)
			time.Sleep(time.Second * 5)
		}
	}()
	<-done
}
