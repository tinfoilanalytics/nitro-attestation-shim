package http

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/blocky/nitrite"
	"github.com/stretchr/testify/assert"
	"github.com/tinfoilsh/nitro-attestation-shim/pkg/util"
	"github.com/tinfoilsh/verifier/pkg/attestation"

	"github.com/tinfoilsh/nitro-attestation-shim/pkg/attestation/nitro"
)

func TestServerNitroRemoteAttestation(t *testing.T) {
	attestationProvider, rootCert, err := nitro.NewMockAttester()
	assert.Nil(t, err)

	server, err := New(8080, 0, attestationProvider, []string{})
	assert.Nil(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:8089")
	assert.Nil(t, err)

	cert, err := util.TLSCertificate("localhost")
	assert.Nil(t, err)

	go func() {
		assert.Nil(t, server.listenWith(listener, cert))
	}()
	time.Sleep(250 * time.Millisecond)

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Fetch remote attestation document
	attResp, err := http.Get("https://localhost:8089/.well-known/tinfoil-attestation")
	assert.Nil(t, err)
	assert.Equal(t, attResp.StatusCode, http.StatusCreated)
	certFP := sha256.Sum256(attResp.TLS.PeerCertificates[0].Raw)

	var attDoc attestation.Document
	assert.Nil(t, json.NewDecoder(attResp.Body).Decode(&attDoc))

	cp := x509.NewCertPool()
	cp.AddCert(rootCert)
	attestation.NitroEnclaveVerifierOpts = nitrite.VerifyOptions{
		Roots: cp,
	}
	defer func() {
		attestation.NitroEnclaveVerifierOpts = nitrite.VerifyOptions{}
	}()

	expectedMeasurements := &attestation.Measurement{
		Type: attestation.AWSNitroEnclaveV1,
		Registers: []string{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"0101010101010101010101010101010101010101010101010101010101010101",
			"0202020202020202020202020202020202020202020202020202020202020202",
		},
	}

	measurements, userData, err := attDoc.Verify()
	assert.Nil(t, err)
	assert.Equal(t, expectedMeasurements, measurements)
	assert.Equal(t, userData, certFP[:])
}
