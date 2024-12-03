/*
Based on go-acme/lego per MIT license below.

https://github.com/go-acme/lego/blob/19a02023b4f22680f404add31b9ec5bf9da8935f/challenge/tlsalpn01/tls_alpn_challenge_server.go

The MIT License (MIT)

Copyright (c) 2017-2024 Ludovic Fernandez
Copyright (c) 2015-2017 Sebastian Erhart

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/log"
	"github.com/mdlayher/vsock"
)

// ProviderServer implements ChallengeProvider for `TLS-ALPN-01` challenge.
type ProviderServer struct {
	listener net.Listener
}

// Present generates a certificate with an SHA-256 digest of the keyAuth provided
// as the acmeValidation-v1 extension value to conform to the ACME-TLS-ALPN spec.
func (s *ProviderServer) Present(domain, token, keyAuth string) error {
	// Generate the challenge certificate using the provided keyAuth and domain.
	cert, err := tlsalpn01.ChallengeCert(domain, keyAuth)
	if err != nil {
		return err
	}

	// Place the generated certificate with the extension into the TLS config
	// so that it can serve the correct details.
	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{*cert}

	// We must set that the `acme-tls/1` application level protocol is supported
	// so that the protocol negotiation can succeed. Reference:
	// https://www.rfc-editor.org/rfc/rfc8737.html#section-6.2
	tlsConf.NextProtos = []string{tlsalpn01.ACMETLS1Protocol}

	vsockListener, err := vsock.Listen(cfg.VsockListenPort, nil)
	if err != nil {
		return fmt.Errorf("creating vsock listener for TLS-ALPN-01: %w", err)
	}
	s.listener = tls.NewListener(vsockListener, tlsConf)

	// Shut the server down when we're finished.
	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Println(err)
		}
	}()

	log.Print("TLS listener started. Waiting for vsock proxy initialization...")
	time.Sleep(5 * time.Second)
	log.Println("Done waiting for vsock proxy initialization. Proceeding with certificate request.")

	return nil
}

// CleanUp closes the HTTPS server.
func (s *ProviderServer) CleanUp(domain, token, keyAuth string) error {
	if s.listener == nil {
		return nil
	}

	// Server was created, close it.
	if err := s.listener.Close(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
