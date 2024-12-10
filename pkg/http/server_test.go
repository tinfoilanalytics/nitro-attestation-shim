package http

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/blocky/nitrite"
	"github.com/cloudflare/circl/hpke"
	"github.com/stretchr/testify/assert"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation"
)

func TestServerNitroAttestation(t *testing.T) {
	attestationProvider, cert, err := attestation.NewMockAttester()
	assert.Nil(t, err)
	att, err := NewAttestationConfig(attestationProvider)
	assert.Nil(t, err)

	server, err := New(8080, *att)
	assert.Nil(t, err)
	listener, err := net.Listen("tcp", ":8080")
	assert.Nil(t, err)

	go server.Serve(listener)
	time.Sleep(250 * time.Millisecond)

	cp := x509.NewCertPool()
	cp.AddCert(cert)

	attResp, err := http.Get("http://localhost:8080/.well-known/nitro-attestation")
	assert.Nil(t, err)

	var attDoc string
	assert.Nil(t, json.NewDecoder(attResp.Body).Decode(&attDoc))

	attDocBytes, err := base64.StdEncoding.DecodeString(attDoc)
	assert.Nil(t, err)

	resp, err := nitrite.Verify(attDocBytes, nitrite.VerifyOptions{Roots: cp})
	assert.Nil(t, err)
	assert.Equal(t, resp.Document.ModuleID, "Mock Module")

	pubkey, err := UnmarshalPubKey(server.pubKey)
	assert.Nil(t, err)
	assert.Equal(t, pubkey.Scheme(), hpke.KEM_P384_HKDF_SHA384.Scheme())
}
