package tls

import (
	"fmt"
	"log"
	"net"

	"github.com/mdlayher/vsock"
	"github.com/miekg/dns"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/util"
)

func handleConnection(conn net.Conn, vsockPort uint32) {
	defer conn.Close()
	log.Printf("Accepted connection from %v", conn.RemoteAddr())

	vsockConn, err := vsock.Dial(3, vsockPort, nil)
	if err != nil {
		log.Printf("Failed to connect to vsock port %d: %v", vsockPort, err)
		return
	}
	defer vsockConn.Close()

	log.Printf("Connected to vsock port %d", vsockPort)
	util.CopyBetween(conn, vsockConn)
	log.Printf("Connection closed")
}

func serveDNS() {
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)

		for _, q := range r.Question {
			if q.Qtype == dns.TypeA {
				rr := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					A: net.ParseIP("127.0.0.1"),
				}
				m.Answer = append(m.Answer, rr)
			}
		}

		w.WriteMsg(m)
	})

	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("Starting DNS server on :53")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Failed to start DNS server: %v", err)
	}
}

// Proxy starts a TCP listener and proxies connections to a vsock port
func Proxy(tcpPort, vsockPort uint32) {
	go serveDNS()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		log.Fatalf("Failed to listen on %d: %v", tcpPort, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn, vsockPort)
	}
}
