#!/bin/sh

set -e

ip addr add dev lo 127.0.0.1/32
ip link set lo up

rm -rf /etc/resolv.conf
echo "nameserver 127.0.0.1" >> /etc/resolv.conf

export SHIM_TLS_DOMAIN=tls-enclave-test.tinfoil.sh
export SHIM_TLS_EMAIL=
export SHIM_PORT=6000
export SHIM_UPSTREAM_PORT=8080

/tls-egress-shim &
sleep 1 && /http-ingress-shim &

mkdir /tmp/public
cd /tmp/public
echo "Hello, World!" > index.html
python3 -m http.server 8080
