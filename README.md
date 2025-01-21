# Nitro Enclave Attestation Shim

The Nitro Attestation Shim requests a TLS certificate from Let's Encrypt, terminates TLS within an AWS Nitro Enclave, and serves the remote attestation including the TLS certificate fingerprint.

## Usage

```bash
nitro-attestation-shim -u UPSTREAM_PORT -e EMAIL [OPTIONS] -- [COMMAND]
```

### Required Flags

- `-u, --upstream-port`: HTTP port to connect to upstream server
- `-e, --email`: Email address for Let's Encrypt account registration

### Optional Flags

- `-s, --staging-ca`: Use Let's Encrypt staging environment instead of production
- `-p, --paths`: Specific paths to proxy to the upstream server (if empty, all paths are proxied)

## AF_VSOCK Ports

| Port | Direction | Description                          |
|------|-----------|--------------------------------------|
| 443  | Listen    | Service proxy and attestation server |
| 8080 | Listen    | Control API (HTTP)                   |
| 7443 | Connect   | Host TLS egress proxy                |

## Network Configuration

The shim automatically configures a loopback address. The runtime container must support `iproute2`.

## Example

```bash
nitro-attestation-shim -u 8000 -e admin@example.com -s -p /api/v1 -- python3 -m http.server
```

## Notes

- The shim waits for 1 second on startup to allow for console attachment

## Host preparation on Amazon Linux

```bash
sudo dnf install -y git socat docker aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel
sudo usermod -aG ne ec2-user
sudo usermod -aG docker ec2-user

# Optionally increase memory and CPU allocation
sudo sed -i 's/^memory.*/memory_mib: 24576/' /etc/nitro_enclaves/allocator.yaml

sudo systemctl enable --now docker
sudo systemctl enable --now nitro-enclaves-allocator
```

## Development

To build a new release, push a new tag using semver (`vX.Y.Z`). GitHub Actions will build and publish the image to `ghcr.io/tinfoilanalytics/nitro-attestation-shim`.

The shim container image doesn't run any code itself, but rather serves as a parent layer for the application specific container image. The shim binary is available at `/nitro-attestation-shim` in the container to copy into your runtime layer. See the [nginx Dockerfile](https://github.com/tinfoilanalytics/nitro-attestation-shim/blob/main/examples/nginx/Dockerfile) for a simple example.
