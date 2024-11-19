FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim AS shim

FROM ollama/ollama

COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim

ENV NITRO_SHIM_PORT=6000
ENV NITRO_SHIM_UPSTREAM_PORT=11434

RUN nohup bash -c "ollama serve &" && sleep 5 && ollama pull llama3.2:1b

ENTRYPOINT ["/nitro-attestation-shim", "/bin/ollama", "serve"]
