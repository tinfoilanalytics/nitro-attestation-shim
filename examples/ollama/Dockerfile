FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim

FROM ollama/ollama

COPY --from=shim /shim /nitro-attestation-shim

RUN apt update -y
RUN apt install -y iproute2

ENV HOME=/

RUN nohup bash -c "ollama serve &" && sleep 5 && ollama pull llama3.2:1b

ENTRYPOINT ["/nitro-attestation-shim", "-p", "11434", "--", "/bin/ollama", "serve"]
