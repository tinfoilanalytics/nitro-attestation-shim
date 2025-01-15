package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type User struct {
	Email        string
	Registration *registration.Resource
	key          *ecdsa.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (u *User) GetPublicKeyBytes() []byte {
	publicKeyECDSA, ok := u.key.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil
	}
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKeyECDSA)
	if err != nil {
		return nil
	}
	return x509EncodedPub
}

func NewUser(email string) (*User, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &User{
		Email: email,
		key:   privateKey,
	}, nil
}

var _ registration.User = &User{}

func RequestCertificate(
	domain, ca string,
	user *User,
	vsockListenPort uint32,
) (*tls.Certificate, error) {
	config := lego.NewConfig(user)
	config.CADirURL = ca
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("could not create lego client: %w", err)
	}
	log.Printf("Setting up TLS-ALPN-01 challenge listener")
	if err := client.Challenge.SetTLSALPN01Provider(newProviderServer(vsockListenPort)); err != nil {
		return nil, fmt.Errorf("could not set up TLS-ALPN-01 challenge listener: %w", err)
	}
	log.Printf("Registering User")
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("could not register User: %w", err)
	}
	user.Registration = reg
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("could not obtain certificate: %w", err)
	}
	keypair, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could not create keypair: %w", err)
	}
	return &keypair, nil
}
