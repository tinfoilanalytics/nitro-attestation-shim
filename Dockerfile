FROM scratch
ENTRYPOINT ["/true"]
COPY shim /
COPY tls-egress-shim-host /
