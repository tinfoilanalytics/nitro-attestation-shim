FROM scratch
ENTRYPOINT ["/true"]
COPY http-ingress-shim /
COPY tls-egress-shim /
