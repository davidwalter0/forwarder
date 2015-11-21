package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var counter uint64 = 0
var wgroup *sync.WaitGroup = new(sync.WaitGroup)

func forward(connection net.Conn, n uint64) {
	wgroup.Add(3)
	go func() {
		done := make(chan bool, 2)
		client, err := net.Dial("tcp", os.Args[2])

		if err != nil {
			log.Fatalf("Connection failed: %v", err)
		}
		log.Printf("Connection[%16d] Connect               %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
		go func() {
			defer client.Close()
			defer connection.Close()
			io.Copy(client, connection)
			done <- true
			log.Printf("Connection[%16d] Closing Client Recv   %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
			wgroup.Done()
		}()

		go func() {
			defer client.Close()
			defer connection.Close()
			io.Copy(connection, client)
			done <- true
			log.Printf("Connection[%16d] Closing Client Send   %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
			wgroup.Done()
		}()

		for i := 0; i < 2; i++ {
			<-done
		}
		log.Printf("Connection[%16d] Closed                %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
	}()
	wgroup.Done()
}

var Build string
var Commit string

func main() {
	array := strings.Split(os.Args[0], "/")
	me := array[len(array)-1]
	fmt.Println(me, "version built as:", Build, "commit:", Commit)

	defer wgroup.Wait()
	if len(os.Args) != 3 {
		log.Fatalf("Usage %s frontend-ip:port backend-ip:port\n", os.Args[0])
		return
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("net.Listen(\"tcp\", %s ) failed: %v", os.Args[1], err)
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalf("ERROR: failed to accept listener: %v", err)
		}
		atomic.AddUint64(&counter, 1)
		n := atomic.LoadUint64(&counter)
		log.Printf("Connection[%16d] Open                  %45v %45v\n", n, connection.LocalAddr(), connection.RemoteAddr())
		go forward(connection, n)
	}
}
