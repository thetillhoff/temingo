FROM alpine:3.24.4@sha256:93b41ef40b1677fce8fcc64ff4af09daddfacf5040afde618bc46c1e2d1cc7d4
ARG TARGETOS
ARG TARGETARCH

# Copy the pre-built binary directly from artifacts by name
COPY --chmod=755 artifacts/temingo_${TARGETOS}_${TARGETARCH} /usr/local/bin/temingo

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/temingo"]
