package attestation

import (
	_ "embed"
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

type NitroProvider struct{}

var _ Provider = (*NitroProvider)(nil)

func New() *NitroProvider {
	return &NitroProvider{}
}

func (n *NitroProvider) RequestAttestation(r *request.Attestation) ([]byte, error) {
	sess, err := nsm.OpenDefaultSession()
	if nil != err {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(r)
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
