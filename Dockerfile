FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
ARG TARGETOS
ARG TARGETARCH

# Copy the pre-built binary directly from artifacts by name
COPY --chmod=755 artifacts/temingo_${TARGETOS}_${TARGETARCH} /usr/local/bin/temingo

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/temingo"]
