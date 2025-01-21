package control

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type status string

var (
	statusWaiting    status = "waiting for domain"
	statusRequesting status = "requesting certificate"
	statusReady      status = "ready"
)

type Server struct {
	status status
	mux    *http.ServeMux

	Ready chan struct{}
}

func New(requestCertCallback func(string) error) *Server {
	s := &Server{
		status: statusWaiting,
		mux:    http.NewServeMux(),
		Ready:  make(chan struct{}),
	}

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("nitro-attestation-shim control server"))
	})

	s.mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusOK, map[string]string{
			"status": string(s.status),
		})
	})

	s.mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		if s.status != statusWaiting {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "domain already set"})
			return
		}

		domain := r.URL.Query().Get("domain")
		if domain == "" {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "missing domain"})
			return
		}

		s.status = statusRequesting
		if err := requestCertCallback(domain); err != nil {
			sendJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		s.Ready <- struct{}{}
		s.status = statusReady
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return s
}

func sendJSON(w http.ResponseWriter, httpCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal JSON: %s", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Listen(listener net.Listener) error {
	return http.Serve(listener, s.mux)
}
