package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
	"github.com/mdlayher/vsock"
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

func getInt(env string, defaultValue int) int {
	val, err := strconv.Atoi(os.Getenv(env))
	if nil != err {
		return defaultValue
	}
	return val
}

func linkUp() error {
	if err := exec.Command("ip", "addr", "add", "dev", "lo", "127.0.0.1/32").Run(); err != nil {
		return err
	}
	return exec.Command("ip", "link", "set", "lo", "up").Run()
}

func main() {
	listenPort := getInt("NITRO_SHIM_PORT", 6000)
	upstreamHost := fmt.Sprintf("localhost:%d", getInt("NITRO_SHIM_UPSTREAM_PORT", 6001))
	useVsock := os.Getenv("NITRO_SHIM_LOCAL") == ""
	tlsCertFile := os.Getenv("NITRO_SHIM_TLS_CERT")
	tlsKeyFile := os.Getenv("NITRO_SHIM_TLS_KEY")

	if useVsock {
		if err := linkUp(); err != nil {
			log.Fatalf("setting up lo failed: %s", err)
			return
		}
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command> [args...]", os.Args[0])
	}

	var l net.Listener
	var err error
	if useVsock {
		l, err = vsock.Listen(uint32(listenPort), nil)
	} else {
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	}
	if err != nil {
		log.Fatalf("listen: %s", err)
		return
	}
	defer l.Close()

	http.HandleFunc("/.well-known/nitro-attestation", func(w http.ResponseWriter, r *http.Request) {
		att, err := requestAttestation(&request.Attestation{})
		if nil != err {
			http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(att)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   upstreamHost,
		})
		proxy.ServeHTTP(w, r)
	})

	go func() {
		if tlsCertFile != "" && tlsKeyFile != "" {
			log.Printf("Starting HTTPS server on port %d\n", listenPort)
			log.Fatal(http.ServeTLS(l, nil, tlsCertFile, tlsKeyFile))
		} else {
			log.Printf("Starting HTTP server on port %d\n", listenPort)
			log.Fatal(http.Serve(l, nil))
		}
	}()

	log.Printf("Running command: %s\n", os.Args[1:])
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("command failed with %s\n", err)
	}
}
