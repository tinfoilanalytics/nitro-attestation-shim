FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim:v0.1.3 AS shim
FROM nginxdemos/hello

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim

ENTRYPOINT ["/nitro-attestation-shim", "-u", "80", "-c", "7443", "-l", "6000", "--", "nginx", "-g", "daemon off;"]
