package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	"github.com/mdlayher/vsock"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/http"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/tls"
)

var version = "dev" // set by the build system

var opts struct {
	Ports            []uint32 `short:"p" description:"list of HTTP ports to expose to the enclave"`
	HostTLSProxyPort uint32   `long:"c" description:"vsock port to connect to host side proxy"`
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
	log.SetPrefix("[nitro-attestation-shim] ")
	log.Printf("Version: %s", version)

	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalf("parsing flags: %s", err)
	}

	if opts.HostTLSProxyPort == 0 {
		opts.HostTLSProxyPort = 7443
	}

	if err := setupNetworking(); err != nil {
		log.Fatalf("configuring container networking: %s", err)
	}

	att, err := http.NewAttestationConfig(attestation.New())
	if err != nil {
		log.Fatalf("creating attestation config: %s", err)
	}

	for _, port := range opts.Ports {
		srv, err := http.New(port, *att)
		if err != nil {
			log.Fatalf("creating HTTP server: %s", err)
		}

		listener, err := vsock.Listen(port, nil)
		if err != nil {
			log.Fatalf("creating vsock listener: %s", err)
		}

		log.Printf("Starting HTTP server on vsock:%d", port)
		go func() {
			log.Fatal(srv.Serve(listener))
		}()
		log.Printf("HTTP server on vsock:%d started", port)
	}

	var tcpPort uint32 = 443
	log.Printf("Listening on %d, proxying to vsock port %d", tcpPort, opts.HostTLSProxyPort)
	go tls.Proxy(tcpPort, opts.HostTLSProxyPort)

	// Run command
	cmd := exec.Command(args[1], args[2:]...)
	log.Printf("Running command: %v\n", cmd.Args)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Fatal(cmd.Run())
}
