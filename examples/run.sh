#!/bin/bash
set -ex

EXAMPLE=$1

sudo pkill socat || true
sudo nitro-cli terminate-enclave --all

sudo nitro-cli run-enclave \
    --cpu-count 2 \
    --memory 2048 \
    --eif-path "$EXAMPLE/$EXAMPLE.eif" \
    --enclave-cid 16

sudo nitro-cli describe-enclaves

sudo socat tcp-listen:443,reuseaddr,fork "vsock-connect:$(nitro-cli describe-enclaves | jq -r '.[0].EnclaveCID'):443" &

sudo nitro-cli console --enclave-name "$EXAMPLE"
