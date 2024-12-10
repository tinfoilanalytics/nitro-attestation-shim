FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim
FROM golang:1.23.3 AS build

COPY main.go .
RUN CGO_ENABLED=0 go build -o /app main.go

FROM alpine:3

RUN apk add iproute2

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim
COPY --from=build /app /app

ENTRYPOINT ["/nitro-attestation-shim", "-p", "80", "-c", "7443", "--", "/app"]
