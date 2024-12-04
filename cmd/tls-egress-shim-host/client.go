package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/mdlayher/vsock"
)

func DialTLSContext(ctx context.Context, _, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Printf("Failed to split host and port: %v", err)
		return nil, err
	}

	conn, err := vsock.DialContext(ctx, 3, 7443, nil)
	if err != nil {
		log.Printf("Failed to connect to target: %v", err)
		return nil, err
	}

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
	})
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("Failed to handshake: %v", err)
		return nil, err
	}

	return tlsConn, nil
}
