package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/caarlos0/env/v11"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
	"github.com/mdlayher/vsock"
)

var (
	version = "dev"
	pubkey  []byte
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
	UpstreamPort    int    `env:"UPSTREAM_PORT" envDefault:"8080"`
	TLSDomain       string `env:"TLS_DOMAIN"`
	TLSEmail        string `env:"TLS_EMAIL"`
}

func main() {
	log.SetPrefix("[http-ingress-shim] ")

	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "SHIM_"}); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}
	log.Printf("version %s: %+v\n", version, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/nitro-attestation", func(w http.ResponseWriter, r *http.Request) {
		att, err := requestAttestation(&request.Attestation{
			PublicKey: pubkey,
		})
		if nil != err {
			http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(att)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", cfg.UpstreamPort),
		})
		proxy.ServeHTTP(w, r)
	})

	if cfg.TLSDomain != "" {
		log.Printf("Requesting certificate for domain %s\n", cfg.TLSDomain)

		user, err := newUser(cfg.TLSEmail)
		if err != nil {
			log.Fatalf("failed to create User: %s", err)
		}
		pubkey = user.GetPublicKeyBytes()

		certs, err := requestCertificate(cfg.TLSDomain, user)
		if err != nil {
			log.Fatalf("failed to get certificate: %s", err)
		}

		server := &http.Server{
			Handler: mux,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{*certs},
				MinVersion:   tls.VersionTLS12,
			},
		}

		log.Printf("Starting HTTPS server on port %d\n", cfg.VsockListenPort)
		l, err := vsock.Listen(cfg.VsockListenPort, nil)
		if err != nil {
			log.Fatalf("creating outer vsock listener: %s", err)
			return
		}
		defer l.Close()
		log.Fatal(server.ServeTLS(l, "", ""))
	} else {
		log.Printf("Starting HTTP server on port %d\n", cfg.VsockListenPort)
		l, err := vsock.Listen(cfg.VsockListenPort, nil)
		if err != nil {
			log.Fatalf("outer vsock listener: %s", err)
			return
		}
		defer l.Close()
		log.Fatal(http.Serve(l, mux))
	}
}
