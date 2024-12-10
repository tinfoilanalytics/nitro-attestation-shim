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

func main() {
	log.SetPrefix("[nitro-attestation-shim] ")
	log.Printf("Version: %s", version)

	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalf("parsing flags: %s", err)
	}

	att, err := http.NewAttestationConfig(attestation.New())
	if err != nil {
		log.Fatalf("creating attestation config: %s", err)
	}

	for _, port := range opts.Ports {
		srv, err := http.New(port, *att)
		listener, err := vsock.Listen(port, nil)
		if err != nil {
			log.Fatalf("creating vsock listener: %s", err)
		}

		log.Printf("Starting HTTP server on vsock:%d", port)
		go log.Fatal(srv.Serve(listener))
	}

	var tcpPort uint32 = 443
	log.Printf("Listening on %d, proxying to vsock port %d", tcpPort, opts.HostTLSProxyPort)
	go tls.Proxy(tcpPort, opts.HostTLSProxyPort)

	// Run command
	fmt.Printf("Running command: %v\n", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Fatal(cmd.Run())
}
