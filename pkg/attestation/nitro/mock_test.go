package nitro

import (
	"crypto/x509"
	"testing"

	"github.com/blocky/nitrite"
	"github.com/stretchr/testify/assert"

	"github.com/tinfoilsh/verifier/attestation"
)

func TestMockAttestation(t *testing.T) {
	attester, rootCert, err := NewMockAttester()
	assert.Nil(t, err)

	expectedUserData := []byte("test user data")

	attDoc, err := attester.RequestAttestation(expectedUserData)
	assert.Nil(t, err)
	assert.NotNil(t, attDoc)

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
	assert.Equal(t, expectedUserData, userData)
}
