package main

import (
	"flag"
	"log"
	"net"

	"github.com/mdlayher/vsock"
	"github.com/miekg/dns"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/util"
)

var version = "dev"

var (
	listenPort = flag.String("port", ":443", "TCP port to listen on")
	vsockPort  = flag.Uint("vsock-port", 7443, "vsock port to connect to")
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted connection from %v", conn.RemoteAddr())

	vsockConn, err := vsock.Dial(3, uint32(*vsockPort), nil)
	if err != nil {
		log.Printf("Failed to connect to vsock port %d: %v", *vsockPort, err)
		return
	}
	defer vsockConn.Close()

	log.Printf("Connected to vsock port %d", *vsockPort)
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

func main() {
	log.SetPrefix("[tls-egress-shim] ")
	log.Printf("version %s\n", version)

	flag.Parse()

	go serveDNS()

	listener, err := net.Listen("tcp", *listenPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *listenPort, err)
	}
	defer listener.Close()

	log.Printf("Listening on %s, proxying to vsock port %d", *listenPort, *vsockPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
