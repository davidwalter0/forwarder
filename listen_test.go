package main

import (
	"fmt"
	// "log"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

type Tester struct {
	t *testing.T
	Pipe
}

type Testers []*Tester

func (f *Pipes) CloseAll() {
	for _, close := range *f {
		close.Close()
	}
}

func TestChannel(t *testing.T) {
	SourceConn := make(chan bool)
	// SinkConn := make(chan bool)
	// SinkConn <- true
	x := 0
	go func() {
		// for {
		select {
		case c := <-SourceConn:
			fmt.Println(c)
			x++
			// case c := <-SinkConn:
			// 	fmt.Println(c)
			// 	x++
			// 	close(SourceConn)
		}
		// if x == 2 {
		// 	break
		// }
		// }
	}()
	close(SourceConn)
	// close(SinkConn)
	// SourceConn <- true
}

func (tester *Tester) Listen() {
	message := "func (tester *Tester) Listen()\n"
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		tester.t.Fatal(err)
	}
	defer CheckClose(conn)
	tester.Pipe.SourceConn = conn

	if _, err := fmt.Fprintf(conn, message); err != nil {
		tester.t.Fatal(err)
	}

}

func TestConn(t *testing.T) {
	message := "Hi there!\n"

	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		if _, err := fmt.Fprintf(conn, message); err != nil {
			t.Fatal(err)
		}
	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(buf[:]))
		if msg := string(buf[:]); msg != message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", msg, message)
		}
		return // Done
	}

}
