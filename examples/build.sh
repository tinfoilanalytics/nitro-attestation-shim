#!/bin/bash

EXAMPLE=$1

cd $EXAMPLE && docker build -t "$EXAMPLE-nitro" .
nitro-cli build-enclave \
    --docker-uri "$EXAMPLE-nitro" \
    --output-file "$EXAMPLE.eif"
