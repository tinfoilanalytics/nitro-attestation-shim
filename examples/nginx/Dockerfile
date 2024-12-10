FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim
FROM nginxdemos/hello

COPY --from=shim /shim /nitro-attestation-shim

ENTRYPOINT ["/nitro-attestation-shim", "nginx", "-g", "daemon off;"]
