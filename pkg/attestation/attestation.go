package attestation

import "github.com/hf/nsm/request"

type Provider interface {
	RequestAttestation(request *request.Attestation) ([]byte, error)
}
