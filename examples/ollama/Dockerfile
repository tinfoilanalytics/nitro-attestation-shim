FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim:v0.1.2 AS shim

FROM ollama/ollama

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim

RUN apt update -y
RUN apt install -y iproute2

ENV HOME=/

RUN nohup bash -c "ollama serve &" && sleep 5 && ollama pull llama3.2:1b

ENTRYPOINT ["/nitro-attestation-shim", "-s", "-d", "inference-enclave.tinfoil.sh", "-e", "nate@tinfoil.sh", "-u", "11434", "-c", "7443", "-l", "443", "--", "/bin/ollama", "serve"]
