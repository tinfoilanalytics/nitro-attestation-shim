package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mdlayher/vsock"
)

func main() {
	transport := http.DefaultTransport.(*http.Transport)
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := vsock.Dial(3, 7000, nil)
		if err != nil {
			log.Printf("Failed to connect to target: %v", err)
			return nil, err
		}
		return conn, nil
	}

	for {
		resp, err := http.Get("http://ipinfo.io/json")
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		log.Printf("Response: %s", body)

		time.Sleep(1 * time.Second)
	}
}
