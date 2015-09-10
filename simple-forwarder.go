package main

import (
    "io"
    "log"
    "net"
    "os"
)

func forward( connection net.Conn ) {
    go func() {
        done := make(chan bool, 2)
        client, err := net.Dial("tcp", os.Args[2])

        if err != nil {
            log.Fatalf("Connection failed: %v", err)
        }
        log.Printf("Connect   %24v %24v\n", connection.LocalAddr(), connection.RemoteAddr() )

        go func() {
            defer client.Close()
            defer connection.Close()
            io.Copy(client, connection)
            done <- true
        }()

        go func() {
            defer client.Close()
            defer connection.Close()
            io.Copy(connection, client)
            done <- true
        }()

        for i := 0; i < 2; i++ {
            <-done
            log.Printf("Received  %24v %24v %d\n", connection.LocalAddr(), connection.RemoteAddr(), i )
        }
        log.Printf("Closed    %24v %24v\n", connection.LocalAddr(), connection.RemoteAddr() )
    }()
}

func main() {
    if len(os.Args) != 3 {
        log.Fatalf("Usage %s frontend-ip:port backend-ip:port\n", os.Args[0]);
        return
    }    

    listener, err := net.Listen("tcp", os.Args[1])
    if err != nil {
        log.Fatalf("net.Listen(\"tcp\", %s ) failed: %v", os.Args[1], err )
    }

    for {
        connection, err := listener.Accept()
        if err != nil {
            log.Fatalf("ERROR: failed to accept listener: %v", err)
        }
        log.Printf("Open      %24v %24v\n", connection.LocalAddr(), connection.RemoteAddr() )
        go forward(connection)
    }
}
