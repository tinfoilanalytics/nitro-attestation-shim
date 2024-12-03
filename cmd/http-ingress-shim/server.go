package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/hf/nsm/request"
	"github.com/mdlayher/vsock"
)

type server struct {
	vsockListenPort, httpUpstreamPort uint32

	mux    *http.ServeMux
	cert   *tls.Certificate
	pubKey []byte
}

func newServer(vsockListenPort, httpUpstreamPort uint32) *server {
	s := &server{
		vsockListenPort:  vsockListenPort,
		httpUpstreamPort: httpUpstreamPort,

		mux: http.NewServeMux(),
	}

	s.mux.HandleFunc("/.well-known/nitro-attestation", func(w http.ResponseWriter, r *http.Request) {
		att, err := requestAttestation(&request.Attestation{
			PublicKey: s.pubKey,
		})
		if nil != err {
			http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(att)
	})

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", httpUpstreamPort),
		})
		proxy.ServeHTTP(w, r)
	})

	return s
}

func (s *server) requestCert(domain, email string) error {
	user, err := newUser(email)
	if err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}

	s.cert, err = requestCertificate(domain, user, s.vsockListenPort)
	if err != nil {
		return fmt.Errorf("failed to request certificate: %s", err)
	}

	s.pubKey, err = x509.MarshalPKIXPublicKey(s.cert.Leaf.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %s", err)
	}

	return nil
}

func (s *server) listenTLS() error {
	httpServer := &http.Server{
		Handler: s.mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*s.cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	l, err := vsock.Listen(s.vsockListenPort, nil)
	if err != nil {
		return fmt.Errorf("creating outer vsock listener: %s", err)
	}
	defer l.Close()

	return httpServer.ServeTLS(l, "", "")
}
