FROM scratch
ENTRYPOINT ["/true"]
COPY shim /nitro-attestation-shim
COPY tls-egress-shim-host /tls-egress-shim-host
