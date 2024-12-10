package http

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server struct {
	AttestationConfig
	httpUpstreamPort uint32
	mux              *http.ServeMux
}

// New creates a new HTTP shim server
func New(httpUpstreamPort uint32, att AttestationConfig) (*Server, error) {
	s := &Server{
		AttestationConfig: att,
		httpUpstreamPort:  httpUpstreamPort,
		mux:               http.NewServeMux(),
	}

	s.mux.HandleFunc("/.well-known/nitro-attestation", s.AttestationConfig.handleAttestation)
	s.mux.HandleFunc("/", s.handleProxy)

	return s, nil
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", s.httpUpstreamPort),
	})
	proxy.ServeHTTP(w, r)
}

func (s *Server) Serve(l net.Listener) error {
	return http.Serve(l, s.mux)
}
