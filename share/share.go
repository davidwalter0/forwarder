package share

import (
	"fmt"
	"log"
	"os"
	"time"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/davidwalter0/forwarder/chanqueue"
	"github.com/davidwalter0/go-cfg"
)

const (
	// TickDelay delay between log entries
	TickDelay = time.Duration(10 * time.Second)
	// Open : State
	Open = iota
	// Closed : State
	Closed
)

var Queue = chanqueue.NewChanQueue()

type EnvCfg interface {
	Read()
}

func NewServerCfg() *ServerCfg {
	return new(ServerCfg)
}

func NewClientCfg() *ClientCfg {
	return new(ClientCfg)
}

type ForwarderCfg struct {
	File       string `json:"file"        doc:"yaml format pipe definitions see README.md"       default:"/var/lib/forwarder/pipes.yaml"`
	Debug      bool   `json:"debug"       doc:"increase verbosity"                               default:"false"`
	Kubeconfig string `json:"kubeconfig"  doc:"kubernetes auth secrets / configuration file"     default:""`
	Kubernetes bool   `json:"kubernetes"  doc:"use kubernetes dynamic endpoints from service/ns" default:"true"`
}

// ClientCfg options to configure endPtDefn
type ClientCfg struct {
	TLSCfg
}

// ServerCfg options to configure endPtDefn
type ServerCfg struct {
	ForwarderCfg
	TLSCfg
}

type TLSCfg struct {
	TLS          bool   `json:"tls"           doc:"Connection uses TLS if true, else plain TCP"    default:"false"`
	HostOverride string `json:"host-override" doc:"TLS handshake host to verify with override"     default:"example.com"`
	CaFile       string `json:"ca-file"       doc:"The Root CA file"                               default:"certs/RootCA.crt"`
	CertFile     string `json:"cert-file"     doc:"The TLS cert file"                              default:"certs/example.com.crt"`
	KeyFile      string `json:"key-file"      doc:"The TLS key file"                               default:"certs/example.com.key"`
	ServerAddr   string `json:"server-addr"   doc:"The server address in the format of host:port"  default:"0.0.0.0:10000"`
}

// Read from env variables or command line flags
func (envCfg *ServerCfg) Read() {
	envCfg.ForwarderCfg.Read()
	envCfg.TLSCfg.Read()
	cfg.Finalize()
	if len(envCfg.File) == 0 {
		fmt.Println("Error: Environment setup yaml forward definitions file not set")
		cfg.Usage()
		os.Exit(1)
	}
}

// Read from env variables or command line flags
func (envCfg *ClientCfg) Read() {
	envCfg.TLSCfg.Read()
	cfg.Finalize()
}

// Read from env variables or command line flags
func (envCfg *ForwarderCfg) Read() {
	var err error
	if err = cfg.AddStruct(envCfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// Read from environment variables or command line flags and load
// the env configuration file endPtDefing pairs.
func (envCfg *TLSCfg) Read() {
	var err error
	if err = cfg.AddStruct(envCfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// LoadClientCreds from disk and initialize TLS credentials
func (envCfg *TLSCfg) LoadClientCreds() credentials.TransportCredentials {
	certificate, err := tls.LoadX509KeyPair(
		envCfg.CertFile,
		envCfg.KeyFile,
	)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(envCfg.CaFile)
	if err != nil {
		log.Fatalf("failed to read ca cert: %s", err)
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatal("failed to append certs")
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   envCfg.HostOverride,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})
	return transportCreds
}

// LoadServerCreds return grpc.ServerOption
func (envCfg *TLSCfg) LoadServerCreds() grpc.ServerOption {

	certificate, err := tls.LoadX509KeyPair(
		envCfg.CertFile,
		envCfg.KeyFile,
	)

	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(envCfg.CaFile)
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
