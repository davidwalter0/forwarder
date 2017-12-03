package main

import (
	"flag"
	"fmt"
	// "io"
	"log"
	// "math/rand"
	// "time"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/davidwalter0/forwarder/proto/pipe"
	empty "github.com/golang/protobuf/ptypes/empty"
	// ts "github.com/golang/protobuf/ptypes/timestamp"
)

var (
	withTls            = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca-file", "certs/RootCA.crt", "The file containing the CA root cert file")
	clientCert         = flag.String("cert-file", "certs/example.com.crt", "The client certificate file")
	clientKey          = flag.String("key-file", "certs/example.com.key", "The client key file")
	serverAddr         = flag.String("server-addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server-host-override", "example.com", "The server name use to verify the hostname returned by TLS handshake")
)

// runPipeLogClient connects to pipe log service and monitors the logs
func runPipeLogClient(client pb.WatcherClient) {
	// fmt.Println("Begin again")
	// defer fmt.Println("End again")
	stream, err := client.Watch(context.Background(), &empty.Empty{}) // ignore *empty.Empty)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}
	var row uint64
	for {
		if pipe, err := stream.Recv(); err != nil {
			log.Printf("%v.Recv() got error %v, want %v\n", stream, err, nil)
			break
		} else {
			fmt.Println(pipe.ToString(row))
			row++
		}
	}
}

func loadCreds() credentials.TransportCredentials {
	certificate, err := tls.LoadX509KeyPair(
		*clientCert,
		*clientKey,
	)

	fmt.Println(*clientKey, *clientCert)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(*caFile)
	if err != nil {
		log.Fatalf("failed to read ca cert: %s", err)
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatal("failed to append certs")
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   *serverHostOverride,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})
	return transportCreds
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *withTls {
		creds := loadCreds()
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewWatcherClient(conn)
	for {
		runPipeLogClient(client)
	}
}
