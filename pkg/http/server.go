package http

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mdlayher/vsock"
	log "github.com/sirupsen/logrus"
	"github.com/tinfoilanalytics/verifier/pkg/attestation"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/http/acme"
)

type Server struct {
	domain, email, ca string
	vsockListenPort   uint32
	httpUpstreamPort  uint32

	mux          *http.ServeMux
	ap           *attestation.Provider
	proxiedPaths []string

	cert *tls.Certificate
}

type Metadata struct {
	Domain string `json:"domain"`
}

// New creates a new HTTP shim server
func New(
	domain, email, ca string,
	httpUpstreamPort, vsockListenPort uint32,
	ap attestation.Provider,
	proxiedPaths []string,
) (*Server, error) {
	s := &Server{
		domain: domain,
		email:  email,
		ca:     ca,

		vsockListenPort:  vsockListenPort,
		httpUpstreamPort: httpUpstreamPort,

		mux:          http.NewServeMux(),
		ap:           &ap,
		proxiedPaths: proxiedPaths,
	}

	s.mux.HandleFunc("/.well-known/tinfoil-attestation", s.handleAttestation)
	s.mux.HandleFunc("/.well-known/tinfoil-metadata", s.handleMeta)
	s.mux.HandleFunc("/", s.handleProxy)

	return s, nil
}

func cors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	cors(w, r)

	log.Infof("Request: %s", r.URL.Path)

	if len(s.proxiedPaths) > 0 {
		allowed := false
		for _, path := range s.proxiedPaths {
			if r.URL.Path == path {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "shim: 403", http.StatusForbidden)
			return
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", s.httpUpstreamPort),
	})
	proxy.ServeHTTP(w, r)
}

func (s *Server) handleAttestation(w http.ResponseWriter, r *http.Request) {
	cors(w, r)

	certFP := sha256.Sum256(s.cert.Certificate[0])
	att, err := (*s.ap).RequestAttestation(certFP[:])
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(att); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode attestation: %s", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	cors(w, r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(
		&Metadata{
			Domain: s.domain,
		},
	); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode meta: %s", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) RequestCert() error {
	user, err := acme.NewUser(s.email)
	if err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}

	s.cert, err = acme.RequestCertificate(s.domain, s.ca, user, s.vsockListenPort)
	if err != nil {
		return fmt.Errorf("failed to request certificate: %s", err)
	}

	return nil
}

// listen starts a TLS server on a given listener
func (s *Server) listen(l net.Listener) error {
	if s.cert == nil {
		return fmt.Errorf("server certificate is nil")
	}

	httpServer := &http.Server{
		Handler: s.mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*s.cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	return httpServer.ServeTLS(l, "", "")
}

// Listen starts a TLS server on the configured server's vsock port
func (s *Server) Listen() error {
	l, err := vsock.Listen(s.vsockListenPort, nil)
	if err != nil {
		return fmt.Errorf("creating outer vsock listener: %s", err)
	}
	defer l.Close()

	return s.listen(l)
}

// listenWith starts a TLS server for testing with a given listener and server certificate
func (s *Server) listenWith(l net.Listener, cert *tls.Certificate) error {
	s.cert = cert
	return s.listen(l)
}
