package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-acme/lego/v4/lego"
	"github.com/jessevdk/go-flags"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation/nitro"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/http"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/tls"
)

var version = "dev" // set by the build system

var opts struct {
	HostTLSProxyPort uint32   `short:"c" description:"vsock port to connect to host side proxy"`
	UpstreamPort     uint32   `short:"u" description:"HTTP port to connect to upstream server"`
	VSockListenPort  uint32   `short:"l" description:"vsock port to listen onn"`
	Domain           string   `short:"d" description:"TLS domain (include wildcard prefix to request a random subdomain)"`
	Email            string   `short:"e" description:"TLS account email"`
	StagingCA        bool     `short:"s" description:"Use staging CA"`
	ProxiedPaths     []string `short:"p" description:"Paths to proxy to the upstream server (all if empty)"`
}

func setupNetworking() error {
	if err := exec.Command("ip", "addr", "add", "dev", "lo", "127.0.0.1/32").Run(); err != nil {
		return fmt.Errorf("setting up loopback: %w", err)
	}
	if err := exec.Command("ip", "link", "set", "lo", "up").Run(); err != nil {
		return fmt.Errorf("setting up loopback: %w", err)
	}

	return nil
}

func main() {
	time.Sleep(1 * time.Second) // Startup delay to allow console attach

	log.SetPrefix("[nitro-attestation-shim] ")
	log.Printf("Version: %s", version)

	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalf("parsing flags: %s", err)
	}

	if err := setupNetworking(); err != nil {
		log.Fatalf("configuring container networking: %s", err)
	}

	var tcpPort uint32 = 443
	log.Printf("Listening on %d, proxying to vsock port %d", tcpPort, opts.HostTLSProxyPort)
	go tls.Proxy(tcpPort, opts.HostTLSProxyPort)

	domain, err := http.ParseDomain(opts.Domain)
	if err != nil {
		log.Fatalf("parsing domain: %s", err)
	}

	ca := lego.LEDirectoryProduction
	if opts.StagingCA {
		ca = lego.LEDirectoryStaging
	}

	srv, err := http.New(
		domain, opts.Email, ca,
		opts.UpstreamPort, opts.VSockListenPort,
		nitro.New(), opts.ProxiedPaths,
	)
	if err != nil {
		log.Fatalf("creating HTTP server: %s", err)
	}

	log.Printf("Requesting TLS certificate for %s on behalf of %s from %s", domain, opts.Email, ca)
	if err := srv.RequestCert(); err != nil {
		log.Fatalf("requesting TLS certificate: %s", err)
	}

	go func() {
		log.Printf("Running command: %v\n", args[1:])
		cmd := exec.Command(args[1], args[2:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Fatal(cmd.Run())
	}()

	log.Printf("Starting HTTPS server on vsock:%d", opts.VSockListenPort)
	log.Fatal(srv.Listen())
}
