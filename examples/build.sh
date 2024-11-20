#!/bin/bash

EXAMPLE=$1

docker build -t "$EXAMPLE-nitro" .
nitro-cli build-enclave \
    --docker-uri "$EXAMPLE-nitro" \
    --output-file "$EXAMPLE-nitro.eif"
