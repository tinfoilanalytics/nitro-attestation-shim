package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-acme/lego/v4/lego"
	"github.com/jessevdk/go-flags"
	"github.com/mdlayher/vsock"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation/nitro"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/control"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/http"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/tls"
)

var version = "dev" // set by the build system

const (
	hostProxyPort = 7443
	controlPort   = 8080
	tlsPort       = 443
)

var opts struct {
	UpstreamPort uint32   `short:"u" long:"upstream" description:"HTTP port to connect to upstream server" required:"true"`
	Email        string   `short:"e" long:"email" description:"TLS account email" required:"true"`
	StagingCA    bool     `short:"s" long:"staging-ca" description:"Use staging CA"`
	ProxiedPaths []string `short:"p" long:"paths" description:"Paths to proxy to the upstream server (all if empty)"`
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

	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalf("parsing flags: %s", err)
	}

	log.SetPrefix("[nitro-attestation-shim] ")
	log.Printf("Version: %s", version)

	if err := setupNetworking(); err != nil {
		log.Fatalf("configuring container networking: %s", err)
	}

	log.Printf("Starting local TLS proxy towards host vsock:%d", hostProxyPort)
	go tls.Proxy(443, hostProxyPort)

	srv, err := http.New(
		opts.UpstreamPort, tlsPort,
		nitro.New(), opts.ProxiedPaths,
	)
	if err != nil {
		log.Fatalf("creating HTTP server: %s", err)
	}

	ca := lego.LEDirectoryProduction
	if opts.StagingCA {
		ca = lego.LEDirectoryStaging
	}

	log.Println("Starting control server")
	controlServer := control.New(func(domain string) error {
		log.Printf("Requesing TLS certificate for %s on behalf of %s from %s", domain, opts.Email, ca)
		return srv.RequestCert(domain, opts.Email, ca)
	})
	go func() {
		controlListener, err := vsock.Listen(controlPort, nil)
		if err != nil {
			log.Fatalf("listening on control server: %s", err)
		}
		if err := controlServer.Listen(controlListener); err != nil {
			log.Fatalf("starting control server: %s", err)
		}
	}()

	go func() {
		log.Printf("Running command: %v\n", args[1:])
		cmd := exec.Command(args[1], args[2:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Fatal(cmd.Run())
	}()

	log.Println("Waiting for control server to signal ready")
	<-controlServer.Ready

	log.Printf("Starting HTTPS server on vsock:%d", tlsPort)
	log.Fatal(srv.Listen())
}
