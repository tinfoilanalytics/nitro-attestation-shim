package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"

	"github.com/mdlayher/vsock"
)

var (
	listenPort = flag.Int("port", 7000, "vsock port to listen on")
	targetHost = flag.String("target", "127.0.0.1:8080", "target host:port to proxy to")
)

func proxyBetween(conn1, conn2 net.Conn) {
	defer conn1.Close()
	defer conn2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(conn2, conn1)
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn1, conn2)
	}()

	wg.Wait()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted connection from %v", conn.RemoteAddr())

	target, err := net.Dial("tcp", *targetHost)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", *targetHost, err)
		return
	}

	log.Printf("Connected to target %s", *targetHost)
	proxyBetween(conn, target)
	log.Printf("Connection closed")
}

func main() {
	flag.Parse()

	log.Printf("Starting vsock proxy server on port %d", *listenPort)
	log.Printf("Proxying connections to %s", *targetHost)

	listener, err := vsock.Listen(uint32(*listenPort), nil)
	if err != nil {
		log.Fatalf("Failed to listen on vsock port %d: %v", *listenPort, err)
	}
	defer listener.Close()

	log.Printf("Listening on vsock port %d", *listenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
