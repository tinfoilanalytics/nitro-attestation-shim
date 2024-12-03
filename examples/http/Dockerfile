FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim
FROM alpine:3.20.3

RUN apk add iproute2 curl python3
COPY --from=shim /tls-egress-shim /
COPY --from=shim /http-ingress-shim /

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
