package main

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/mdlayher/vsock"

	"github.com/tinfoilanalytics/nitro-attestation-shim/cmd/http-ingress-shim/server"
	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation"
)

var (
	version = "dev"
)

var cfg struct {
	VsockListenPort uint32 `env:"LISTEN_PORT" envDefault:"6000"`
	UpstreamPort    uint32 `env:"UPSTREAM_PORT" envDefault:"8080"`
}

func main() {
	log.SetPrefix("[http-ingress-shim] ")

	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "SHIM_"}); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}
	log.Printf("version %s: %+v\n", version, cfg)

	srv, err := server.New(cfg.UpstreamPort, attestation.New())
	listener, err := vsock.Listen(cfg.VsockListenPort, nil)
	if err != nil {
		log.Fatalf("creating vsock listener: %s", err)
	}
	defer listener.Close()

	log.Printf("Starting HTTP server on vsock:%d", cfg.VsockListenPort)
	log.Fatal(srv.Serve(listener))
}
