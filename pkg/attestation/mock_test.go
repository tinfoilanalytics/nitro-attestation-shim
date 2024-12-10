package attestation

import (
	"crypto/x509"
	"testing"

	"github.com/blocky/nitrite"
	"github.com/hf/nsm/request"
	"github.com/stretchr/testify/assert"
)

func TestMockAttestation(t *testing.T) {
	attester, rootCert, err := NewMockAttester()
	assert.Nil(t, err)

	req := &request.Attestation{
		PublicKey: []byte("test public key"),
		UserData:  []byte("test user data"),
		Nonce:     []byte("test nonce"),
	}

	attDoc, err := attester.RequestAttestation(req)
	assert.Nil(t, err)
	assert.NotNil(t, attDoc)

	cp := x509.NewCertPool()
	cp.AddCert(rootCert)
	verified, err := nitrite.Verify(attDoc, nitrite.VerifyOptions{
		Roots: cp,
	})
	assert.Nil(t, err)
	assert.NotNil(t, verified)

	assert.Equal(t, req.PublicKey, verified.Document.PublicKey)
	assert.Equal(t, req.UserData, verified.Document.UserData)
	assert.Equal(t, req.Nonce, verified.Document.Nonce)
}
