# Nitro Enclave Attestation Shim

## Installation on Amazon Linux

```bash
sudo dnf install -y git socat docker aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel
sudo usermod -aG ne ec2-user
sudo usermod -aG docker ec2-user

sudo sed -i 's/^memory.*/memory_mib: 24576/' /etc/nitro_enclaves/allocator.yaml

sudo systemctl enable --now docker
sudo systemctl enable --now nitro-enclaves-allocator
```

## Build and Run

```bash
export EXAMPLE_NAME=nginx
docker build -t "$EXAMPLE_NAME-nitro" .
nitro-cli build-enclave --docker-uri "$EXAMPLE_NAME-nitro" --output-file "$EXAMPLE_NAME-nitro".eif
nitro-cli run-enclave --cpu-count 16 --memory 32G --eif-path "$EXAMPLE_NAME-nitro".eif --debug-mode

```

## Start socat forwarder

```bash
sudo socat tcp-listen:80,reuseaddr,fork vsock-connect:$(nitro-cli describe-enclaves | jq -r '.[0].EnclaveCID'):6000
```

## Development

To build a new release, push a new tag using semver (`vX.Y.Z`). GitHub Actions will build and publish the image to `ghcr.io/tinfoilanalytics/nitro-attestation-shim`.

The shim container image doesn't run any code itself, but rather serves as a parent layer for the application specific container image. The shim binary is available at `/nitro-attestation-shim` in the container to copy into your runtime layer. See the [nginx Dockerfile](https://github.com/tinfoilanalytics/nitro-attestation-shim/blob/main/examples/nginx/Dockerfile) for a simple example.


```bash
cd ~/nitro-attestation-shim/ && ~/go/bin/goreleaser release --snapshot --clean
```