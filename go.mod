module github.com/tinfoilanalytics/nitro-attestation-shim

go 1.23.2

require (
	github.com/caarlos0/env/v11 v11.2.2
	github.com/go-acme/lego/v4 v4.20.4
	github.com/hf/nsm v0.0.0-20220930140112-cd181bd646b9
	github.com/mdlayher/vsock v1.2.1
	github.com/miekg/dns v1.1.62
	golang.org/x/crypto v0.28.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/fxamacker/cbor/v2 v2.2.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/mdlayher/vsock => github.com/natesales/vsock v0.0.0-20241202211744-c23986e4659a
