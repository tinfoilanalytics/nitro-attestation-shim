package http

import (
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudflare/circl/hpke"
	"github.com/cloudflare/circl/kem"
	"github.com/hf/nsm/request"

	"github.com/tinfoilanalytics/nitro-attestation-shim/pkg/attestation"
)

var (
	kemID     = hpke.KEM_P384_HKDF_SHA384
	kdfID     = hpke.KDF_HKDF_SHA384
	aeadID    = hpke.AEAD_AES256GCM
	hpkeSuite = hpke.NewSuite(kemID, kdfID, aeadID)
)

type AttestationConfig struct {
	pubKey              []byte
	privKey             crypto.PrivateKey
	attestationProvider attestation.Provider
}

func (a *AttestationConfig) handleAttestation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	att, err := a.attestationProvider.RequestAttestation(&request.Attestation{
		PublicKey: a.pubKey,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to request attestation: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(att)
}

// NewAttestationConfig generates a HPKE keypair and returns an AttestationConfig wrapping the attestation provider and keypair
func NewAttestationConfig(att attestation.Provider) (*AttestationConfig, error) {
	pub, priv, err := kemID.Scheme().GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generating HPKE key pair: %s", err)
	}

	pubBytes, err := pub.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %s", err)
	}

	return &AttestationConfig{
		pubKey:              pubBytes,
		privKey:             priv,
		attestationProvider: att,
	}, nil
}

func UnmarshalPubKey(b []byte) (kem.PublicKey, error) {
	return kemID.Scheme().UnmarshalBinaryPublicKey(b)
}
