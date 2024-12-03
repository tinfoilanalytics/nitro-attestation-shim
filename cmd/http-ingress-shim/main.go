package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

var (
	version = "dev"
)

func requestAttestation(attestationRequest *request.Attestation) ([]byte, error) {
	sess, err := nsm.OpenDefaultSession()
	if nil != err {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(attestationRequest)
	if nil != err {
		return nil, err
	}

	if res.Error != "" {
		return nil, fmt.Errorf("nsm error: %s", res.Error)
	}
	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, fmt.Errorf("no attestation document from nsm")
	}

	return res.Attestation.Document, nil
}

var cfg struct {
	VsockListenPort uint32 `env:"LISTEN_PORT" envDefault:"6000"`
	UpstreamPort    uint32 `env:"UPSTREAM_PORT" envDefault:"8080"`
	TLSDomain       string `env:"TLS_DOMAIN"`
	TLSEmail        string `env:"TLS_EMAIL"`
}

func main() {
	log.SetPrefix("[http-ingress-shim] ")

	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "SHIM_"}); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}
	log.Printf("version %s: %+v\n", version, cfg)

	s := newServer(cfg.VsockListenPort, cfg.UpstreamPort)

	log.Printf("Requesting certificate for %s", cfg.TLSDomain)
	if err := s.requestCert(cfg.TLSDomain, cfg.TLSEmail); err != nil {
		log.Fatalf("failed to request certificate: %s", err)
	}

	log.Printf("Starting HTTPS server on vsock:%d", cfg.VsockListenPort)
	log.Fatal(s.listenTLS())
}
