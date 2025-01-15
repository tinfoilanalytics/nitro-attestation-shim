package nitro

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/blocky/nitrite"
	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/go-cose"

	"github.com/tinfoilanalytics/verifier/pkg/attestation"
)

type MockProvider struct {
	privateKey  crypto.Signer
	certificate []byte
}

var _ attestation.Provider = (*MockProvider)(nil)

func (a *MockProvider) RequestAttestation(userData []byte) (*attestation.Document, error) {
	doc := nitrite.Document{
		ModuleID:  "Mock Module",
		Timestamp: uint64(time.Now().Unix()),
		Digest:    "SHA384",
		PCRs: map[uint][]byte{
			0: bytes.Repeat([]byte{0}, 32),
			1: bytes.Repeat([]byte{1}, 32),
			2: bytes.Repeat([]byte{2}, 32),
		},
		Certificate: a.certificate,
		CABundle:    [][]byte{a.certificate},
		PublicKey:   nil,
		UserData:    userData,
		Nonce:       nil,
	}

	payload, err := cbor.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshaling attestation document: %w", err)
	}

	msg := cose.UntaggedSign1Message{}
	msg.Headers = cose.Headers{
		Protected: cose.ProtectedHeader{
			cose.HeaderLabelAlgorithm: cose.AlgorithmES384,
		},
	}
	msg.Payload = payload

	signer, err := cose.NewSigner(cose.AlgorithmES384, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("creating signer: %w", err)
	}

	if err := msg.Sign(rand.Reader, nil, signer); err != nil {
		return nil, fmt.Errorf("signing message: %w", err)
	}

	body, err := msg.MarshalCBOR()
	if err != nil {
		return nil, fmt.Errorf("marshaling message: %w", err)
	}

	return &attestation.Document{
		Format: attestation.AWSNitroEnclaveV1,
		Body:   base64.StdEncoding.EncodeToString(body),
	}, nil
}

func NewMockAttester() (*MockProvider, *x509.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generating key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Mock Nitro Attestation",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("creating certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing certificate: %w", err)
	}

	return &MockProvider{
		privateKey:  priv,
		certificate: certDER,
	}, cert, nil
}
