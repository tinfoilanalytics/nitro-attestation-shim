package server

import (
	"crypto"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/cloudflare/circl/hpke"
	"github.com/cloudflare/circl/kem"
	"github.com/hf/nsm/request"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation"
)

var (
	kemID     = hpke.KEM_P384_HKDF_SHA384
	kdfID     = hpke.KDF_HKDF_SHA384
	aeadID    = hpke.AEAD_AES256GCM
	hpkeSuite = hpke.NewSuite(kemID, kdfID, aeadID)
)

type Server struct {
	httpUpstreamPort    uint32
	mux                 *http.ServeMux
	pubKey              []byte
	privKey             crypto.PrivateKey
	attestationProvider attestation.Provider
}

func UnmarshalPubKey(b []byte) (kem.PublicKey, error) {
	return kemID.Scheme().UnmarshalBinaryPublicKey(b)
}

// New creates a new HTTP shim server
func New(httpUpstreamPort uint32, attestationProvider attestation.Provider) (*Server, error) {
	pub, priv, err := kemID.Scheme().GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generating HPKE key pair: %s", err)
	}

	pubBytes, err := pub.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %s", err)
	}

	s := &Server{
		httpUpstreamPort:    httpUpstreamPort,
		mux:                 http.NewServeMux(),
		pubKey:              pubBytes,
		privKey:             priv,
		attestationProvider: attestationProvider,
	}

	s.mux.HandleFunc("/.well-known/nitro-attestation", s.handleAttestation)
	s.mux.HandleFunc("/", s.handleProxy)

	return s, nil
}

func (s *Server) handleAttestation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	att, err := s.attestationProvider.RequestAttestation(&request.Attestation{
		PublicKey: s.pubKey,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(att)
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
