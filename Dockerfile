FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
ARG TARGETOS
ARG TARGETARCH

# Copy the pre-built binary directly from artifacts by name
COPY --chmod=755 artifacts/temingo_${TARGETOS}_${TARGETARCH} /usr/local/bin/temingo

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/temingo"]
