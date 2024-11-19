package main

import (
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
	"github.com/vishvananda/netlink"
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
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr("127.0.0.1/32")
	if err != nil {
		return err
	}

	return netlink.AddrAdd(lo, addr)
}

func main() {
	listenPort := getInt("NITRO_SHIM_PORT", 6000)
	upstreamHost := fmt.Sprintf("localhost:%d", getInt("NITRO_SHIM_UPSTREAM_PORT", 6001))
	cid := getInt("NITRO_SHIM_CID", 16)
	useVsock := os.Getenv("NITRO_SHIM_LOCAL") == ""

	if err := linkUp(); err != nil {
		log.Fatalf("setting up lo failed: %s", err)
		return
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command> [args...]", os.Args[0])
	}

	var l net.Listener
	var err error
	if useVsock {
		l, err = vsock.ListenContextID(uint32(cid), uint32(listenPort), nil)
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
		w.Write(att)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   upstreamHost,
		})
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Starting server on port %d\n", listenPort)
	go func() {
		log.Fatal(http.Serve(l, nil))
	}()

	log.Printf("Running command: %s\n", os.Args[1:])
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("command failed with %s\n", err)
	}
}
