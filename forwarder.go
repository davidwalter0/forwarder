// TODO
// [X] Add yaml daemonset config option for environment variable for default file location
// [X] Add volume mount for file
// [X] Add forwards.yaml
// [ ] Add file change monitoring and reload

package main

import (
	"fmt"
	"github.com/davidwalter0/go-cfg"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
)

func main() {
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Println(me, "version built as:", Build, "commit:", Commit)
	Configure()
	Run()
	<-complete
}

var complete = make(chan bool)
var counter uint64 = 0

// Forward bidirectional <-> connection to/from up/down stream
func (f *Forwarder) Forward() {
	client, err := net.Dial("tcp", f.UpStream)
	if err != nil {
		log.Printf("Connection failed: %v\n", err)
		return
	}

	log.Printf("Connection[%16d] Connect      %-45v %-45v\n", f.ID, f.Connection.LocalAddr(), f.Connection.RemoteAddr())
	var closed uint64
	closer := func() {
		atomic.AddUint64(&closed, 1)
		n := atomic.LoadUint64(&counter)
		if n == 1 {
			if err := client.Close(); err != nil {
				log.Printf("Close failed with error: %v\n", err)
			}
			if err := f.Connection.Close(); err != nil {
				log.Printf("Close failed with error: %v\n", err)
			}
		}
	}
	go func() {
		defer closer()
		if _, err = io.Copy(client, f.Connection); err != nil {
			log.Printf("Connection failed: %v\n", err)
		}
	}()
	go func() {
		defer closer()
		if _, err = io.Copy(f.Connection, client); err != nil {
			log.Printf("Connection failed: %v\n", err)
		}
	}()
}

// ListenForOneConnectionPair start a listener waiting for downstream
// connections connecting them to their upstream pair
func ListenForOneConnectionPair(downstream, upstream string) {
	listener, err := net.Listen("tcp", downstream)
	if err != nil {
		log.Fatalf("net.Listen(\"tcp\", %s ) failed: %v", downstream, err)
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}
		atomic.AddUint64(&counter, 1)
		n := atomic.LoadUint64(&counter)
		var forwarder = Forwarder{Connection: connection, ID: n, UpStream: upstream, DownStream: downstream}
		forwarder.Forward()
	}
}

// Run parse the forward pairs and create a listener for each
func Run() {
	for k, v := range forwards {
		var downstream, upstream = v.DownStream, v.UpStream
		fmt.Println("key", k, "\n  downstream:", v.DownStream, "\n  upstream", v.UpStream)
		go ListenForOneConnectionPair(downstream, upstream)
	}
}

// Configure from environment variables or command line flags and load
// the configuration file forwarding pairs.
func Configure() {
	var err error
	forwards = make(Forwards, 0)
	if err = cfg.Process("", &config); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(config.File) == 0 {
		fmt.Println("Error: configuration file not set")
		cfg.Usage()
		os.Exit(1)
	}

	var text = Load(config.File)
	err = yaml.Unmarshal(text, &forwards)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

// Load helper function from file to []byte
func Load(filename string) []byte {
	if len(filename) == 0 {
		panic(fmt.Sprintf("Can't Load() a file with an empty name"))
	}

	var err error
	var text []byte
	if len(filename) > 0 {
		text, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("%v\n", err)
			os.Exit(3)
		}
	}
	return text
}
