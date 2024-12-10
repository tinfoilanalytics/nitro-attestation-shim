FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim
FROM nginxdemos/hello

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim

ENTRYPOINT ["/nitro-attestation-shim", "-p", "80", "--", "nginx", "-g", "daemon off;"]
