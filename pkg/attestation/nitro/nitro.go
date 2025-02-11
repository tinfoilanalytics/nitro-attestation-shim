package nitro

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"

	"github.com/tinfoilsh/verifier/attestation"
)

type Provider struct{}

var _ attestation.Provider = (*Provider)(nil)

func New() *Provider {
	return &Provider{}
}

func (n *Provider) RequestAttestation(userData []byte) (*attestation.Document, error) {
	sess, err := nsm.OpenDefaultSession()
	if nil != err {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(&request.Attestation{
		UserData:  userData,
		Nonce:     nil,
		PublicKey: nil,
	})
	if nil != err {
		return nil, err
	}

	if res.Error != "" {
		return nil, fmt.Errorf("nsm error: %s", res.Error)
	}
	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, fmt.Errorf("no attestation document from nsm")
	}

	return &attestation.Document{
		Format: attestation.AWSNitroEnclaveV1,
		Body:   base64.StdEncoding.EncodeToString(res.Attestation.Document),
	}, nil
}
