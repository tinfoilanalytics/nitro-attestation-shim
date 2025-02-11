module github.com/tinfoilsh/nitro-attestation-shim

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
	github.com/tinfoilsh/verifier v0.0.15
	github.com/veraison/go-cose v1.3.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/google/go-containerregistry v0.20.2 // indirect
	github.com/google/go-sev-guest v0.0.0-00010101000000-000000000000 // indirect
	github.com/google/logger v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/letsencrypt/boulder v0.0.0-20240620165639-de9c06129bec // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/sigstore/sigstore v1.8.9 // indirect
	github.com/theupdateframework/go-tuf/v2 v2.0.1 // indirect
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	marwan.io/wasm-fetch v0.1.0 // indirect
)

replace github.com/mdlayher/vsock => github.com/natesales/vsock v0.0.0-20250115072414-5a011980d3ec

replace github.com/google/go-sev-guest => github.com/jraman567/go-sev-guest v0.0.0-20250117204014-6339110611c9
