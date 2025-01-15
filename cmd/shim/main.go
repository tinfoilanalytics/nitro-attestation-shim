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
	HostTLSProxyPort uint32 `short:"c" description:"vsock port to connect to host side proxy"`
	UpstreamPort     uint32 `short:"u" description:"HTTP port to connect to upstream server"`
	VSockListenPort  uint32 `short:"l" description:"vsock port to listen onn"`
	Domain           string `short:"d" description:"TLS domain"`
	Email            string `short:"e" description:"TLS account email"`
	StagingCA        bool   `short:"s" description:"Use staging CA"`
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

	srv, err := http.New(opts.UpstreamPort, opts.VSockListenPort, nitro.New())
	if err != nil {
		log.Fatalf("creating HTTP server: %s", err)
	}

	ca := lego.LEDirectoryProduction
	if opts.StagingCA {
		ca = lego.LEDirectoryStaging
	}

	log.Printf("Requesting TLS certificate for %s on behalf of %s from %s", opts.Domain, opts.Email, ca)
	if err := srv.RequestCert(opts.Domain, opts.Email, ca); err != nil {
		log.Fatalf("requesting TLS certificate: %s", err)
	}

	log.Printf("Starting HTTPS server on vsock:%d", opts.VSockListenPort)
	go log.Fatal(srv.Listen())

	log.Printf("Running command: %v\n", args[1:])

	// Run command
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Fatal(cmd.Run())
}
