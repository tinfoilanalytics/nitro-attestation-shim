module github.com/tinfoilanalytics/nitro-attestation-shim

go 1.23.2

require (
	github.com/blocky/nitrite v0.0.1
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/go-acme/lego/v4 v4.21.0
	github.com/hf/nsm v0.0.0-20220930140112-cd181bd646b9
	github.com/jessevdk/go-flags v1.6.1
	github.com/mdlayher/vsock v1.2.1
	github.com/miekg/dns v1.1.62
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/tinfoilanalytics/verifier v0.0.0-20250115062353-a98dc5237bc5
	github.com/veraison/go-cose v1.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
)

replace github.com/mdlayher/vsock => github.com/natesales/vsock v0.0.0-20250115072414-5a011980d3ec
