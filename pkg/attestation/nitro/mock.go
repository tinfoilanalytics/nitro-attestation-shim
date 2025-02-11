package nitro

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/blocky/nitrite"
	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/go-cose"

	"github.com/tinfoilsh/verifier/pkg/attestation"

	"github.com/tinfoilsh/nitro-attestation-shim/pkg/util"
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
	priv, certDER, err := util.Certificate("Nitro Mock Certificate")
	if err != nil {
		return nil, nil, err
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
