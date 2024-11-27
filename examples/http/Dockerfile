FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim
FROM golang:1.23.3 AS build

WORKDIR /src
COPY . /src
RUN go build -o /app

FROM scratch

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim
COPY --from=build /app /app

ENV NITRO_SHIM_PORT=6000
ENV NITRO_SHIM_UPSTREAM_PORT=80

ENTRYPOINT ["/nitro-attestation-shim", "/app"]
