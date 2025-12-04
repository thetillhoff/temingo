FROM scratch

# TARGETOS and TARGETARCH are automatically set by Docker Buildx
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Copy the pre-built binary directly from artifacts by name
COPY artifacts/temingo_${TARGETOS}_${TARGETARCH} /usr/local/bin/temingo

ENTRYPOINT ["/usr/local/bin/temingo"]
