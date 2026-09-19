FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
ARG TARGETOS
ARG TARGETARCH

# Copy the pre-built binary directly from artifacts by name
COPY --chmod=755 artifacts/temingo_${TARGETOS}_${TARGETARCH} /usr/local/bin/temingo

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/temingo"]
