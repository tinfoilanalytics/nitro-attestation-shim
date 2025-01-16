#!/bin/bash
set -ex

EXAMPLE=$1

sudo pkill socat || true
sudo nitro-cli terminate-enclave --all

sudo nitro-cli run-enclave \
    --cpu-count 2 \
    --memory 2048 \
    --eif-path "$EXAMPLE/$EXAMPLE.eif" \
    --enclave-cid 16 \
    --debug-mode

sudo nitro-cli describe-enclaves

sudo socat tcp-listen:443,reuseaddr,fork "vsock-connect:16:443" &

sudo nitro-cli console --enclave-name "$EXAMPLE"
