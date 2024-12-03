package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/mdlayher/vsock"
	"gopkg.in/yaml.v3"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/util"
)

var version = "dev"

var (
	configFile = flag.String("config", "config.yml", "path to config file")
)

type Config struct {
	AllowedSNI []string `yaml:"allowed_sni"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &config, nil
}

func handleTLSConnection(conn net.Conn, allowedSNIs []string) {
	defer conn.Close()
	log.Printf("Accepted connection from %v", conn.RemoteAddr())

	buf := new(bytes.Buffer)
	tr := io.TeeReader(conn, buf)

	sni, err := peekSNI(tr)
	if err != nil {
		log.Printf("Error reading SNI: %v", err)
		return
	}
	log.Printf("SNI: %s", sni)

	// Check if SNI is allowed
	allowed := false
	for _, allowedSNI := range allowedSNIs {
		if sni == allowedSNI {
			allowed = true
			break
		}
	}
	if !allowed {
		log.Printf("SNI %s not in allowed list", sni)
		return
	}

	ips, err := net.LookupIP(sni)
	if err != nil {
		log.Printf("Failed to resolve %s: %v", sni, err)
		return
	}
	ip := ips[0].String()

	tcpHost := net.JoinHostPort(ip, "443")
	log.Printf("Dialing %s", tcpHost)
	target, err := net.Dial("tcp", tcpHost)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", sni, err)
		return
	}
	defer target.Close()

	if _, err := io.Copy(target, buf); err != nil {
		log.Printf("Error writing ClientHello: %v", err)
		return
	}

	log.Printf("Connected to target %s (%s)", sni, ip)
	util.CopyBetween(conn, target)
	log.Printf("Connection closed")
}

func peekSNI(r io.Reader) (string, error) {
	// Check if we have a TLS connection
	peek := make([]byte, 4096)
	_, err := io.ReadFull(r, peek[:5])
	if err != nil {
		return "", err
	}
	if peek[0] != 0x16 {
		return "", fmt.Errorf("not a TLS handshake")
	}

	// Read the rest of the ClientHello
	recordLen := int(peek[3])<<8 | int(peek[4])
	if recordLen > 4096-5 {
		recordLen = 4096 - 5
	}
	_, err = io.ReadFull(r, peek[5:recordLen+5])
	if err != nil {
		return "", err
	}

	var hello tls.ClientHelloInfo
	err = tls.Server(readOnlyConn{bytes.NewReader(peek[:recordLen+5])}, &tls.Config{
		GetConfigForClient: func(hi *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = *hi
			return nil, nil
		},
	}).Handshake()

	if hello.ServerName == "" {
		return "", fmt.Errorf("no SNI found")
	}

	return hello.ServerName, nil
}

type readOnlyConn struct {
	r io.Reader
}

func (c readOnlyConn) Read(p []byte) (int, error)         { return c.r.Read(p) }
func (c readOnlyConn) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (c readOnlyConn) Close() error                       { return nil }
func (c readOnlyConn) LocalAddr() net.Addr                { return nil }
func (c readOnlyConn) RemoteAddr() net.Addr               { return nil }
func (c readOnlyConn) SetDeadline(t time.Time) error      { return nil }
func (c readOnlyConn) SetReadDeadline(t time.Time) error  { return nil }
func (c readOnlyConn) SetWriteDeadline(t time.Time) error { return nil }

func main() {
	log.SetPrefix("[tls-egress-shim-server] ")
	log.Printf("version %s\n", version)

	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Loaded config with %d allowed SNIs", len(config.AllowedSNI))

	var port uint32 = 7443

	listener, err := vsock.Listen(port, nil)
	if err != nil {
		log.Fatalf("Failed to listen on vsock port %d: %v", port, err)
	}
	defer listener.Close()

	log.Printf("Listening on vsock port %d", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection on port %d: %v", port, err)
			continue
		}

		go handleTLSConnection(conn, config.AllowedSNI)
	}
}
