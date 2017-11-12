// TODO
// Add yaml daemonset config option for environment variable for default file location
// Add volume mount for file
// Add forwards
// Add file change monitoring and reload

package main

import (
	"encoding/json"
	"fmt"
	"github.com/davidwalter0/go-cfg"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

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
			log.Printf("%v.\n", err)
			os.Exit(3)
		}
	}
	return text
}

type FORWARD_CFG struct {
	File string `json:"mappings file" doc:"yaml format file to import mappings from\n        name:\n          downstream: host:port\n          upstream:   host:port\n        "`
}

var counter uint64 = 0
var wgroup *sync.WaitGroup = new(sync.WaitGroup)

type Forwarder struct {
	Connection net.Conn
	Id         uint64
	UpStream   string
	DownStream string
}

// func forward(connection net.Conn, n uint64) {
func (f *Forwarder) Forward() {
	wgroup.Add(3)
	go func() {
		done := make(chan bool, 2)
		client, err := net.Dial("tcp", f.UpStream)

		if err != nil {
			log.Fatalf("Connection failed: %v", err)
		}
		log.Printf("Connection[%16d] Connect               %45v %45v\n", f.Id, f.Connection.LocalAddr(), f.Connection.RemoteAddr())
		go func() {
			defer client.Close()
			defer f.Connection.Close()
			io.Copy(client, f.Connection)
			done <- true
			log.Printf("Connection[%16d] Closing Recv   %45v %45v\n", f.Id, f.Connection.LocalAddr(), f.Connection.RemoteAddr())
			wgroup.Done()
		}()

		go func() {
			defer client.Close()
			defer f.Connection.Close()
			io.Copy(f.Connection, client)
			done <- true
			log.Printf("Connection[%16d] Closing Send   %45v %45v\n", f.Id, f.Connection.LocalAddr(), f.Connection.RemoteAddr())
			wgroup.Done()
		}()

		for i := 0; i < 2; i++ {
			<-done
		}
		log.Printf("Connection[%16d] Closed           %45v %45v\n", f.Id, f.Connection.LocalAddr(), f.Connection.RemoteAddr())
	}()
	wgroup.Done()
}

var Build string
var Commit string

var config FORWARD_CFG
var forwards Forwards

func Configure() {
	var err error
	forwards = make(Forwards, 0)
	if err = cfg.Parse(&config); err != nil {
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

	var jsonText []byte
	jsonText, err = json.MarshalIndent(&forwards, "", "  ")
	fmt.Printf("jsonText: \n%v\n", string(jsonText))

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("---\nforwards:\n%v\n\n", forwards)
	for k, v := range forwards {
		fmt.Println("key", k, "\n  downstream:", v.DownStream, "\n  upstream", v.UpStream)
	}
	fmt.Printf("---\nforwards:\n%v\n%T\n", forwards, forwards)
}

func Run() {
	for k, v := range forwards {
		var downstream, upstream = v.DownStream, v.UpStream
		fmt.Println("key", k, "\n  downstream:", v.DownStream, "\n  upstream", v.UpStream)

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
			log.Printf("Connection[%16d] Open             %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
			var forwarder = Forwarder{Connection: connection, Id: n, UpStream: upstream, DownStream: downstream}
			go forwarder.Forward()
		}
	}
}

func main() {
	Configure()
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Println(me, "version built as:", Build, "commit:", Commit)

	defer wgroup.Wait()
	Run()
}
