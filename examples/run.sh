#!/bin/bash
set -ex

EXAMPLE=$1

sudo nitro-cli terminate-enclave --all

sudo nitro-cli run-enclave \
    --cpu-count 2 \
    --memory 2048 \
    --eif-path "$EXAMPLE/$EXAMPLE.eif" \
    --enclave-cid 16 \
    --debug-mode

sudo nitro-cli describe-enclaves

sudo nitro-cli console --enclave-name $EXAMPLE
