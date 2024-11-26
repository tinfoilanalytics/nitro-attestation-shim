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
