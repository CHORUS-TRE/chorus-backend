# syntax=docker/dockerfile:1

# Build stage
FROM harbor.build.chorus-tre.ch/docker_proxy/library/ubuntu:24.04 AS builder

# Install build prerequisites
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GOLANG_VERSION=1.24.3
ARG GOLANG_CHECKSUM=3333f6ea53afa971e9078895eaa4ac7204a8c6b5c68c10e6bc9a33e8e391bdd8

RUN curl -fsSL https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz -o go.tar.gz && \
    echo "${GOLANG_CHECKSUM}  go.tar.gz" | sha256sum -c - && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin" \
    GOCACHE="/chorus/.cache/go-build" \
    GOMODCACHE="/chorus/.cache/go-mod" \
    CGO_ENABLED=0

WORKDIR /chorus

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download Go modules with private repository authentication
# This layer will be cached unless go.mod or go.sum changes
RUN --mount=type=cache,target="/chorus/.cache/go-build" \
    --mount=type=cache,target="/chorus/.cache/go-mod" \
    --mount=type=secret,id=GIT_USERNAME \
    --mount=type=secret,id=GIT_PASSWORD \
    if [ -f /run/secrets/GIT_USERNAME ] && [ -f /run/secrets/GIT_PASSWORD ]; then \
        echo "Fetching private dependencies..." && \
        u="$(cat /run/secrets/GIT_USERNAME)" && \
        p="$(cat /run/secrets/GIT_PASSWORD)" && \
        GOPRIVATE=github.com/CHORUS-TRE/* \
        GIT_CONFIG_COUNT=1 \
        GIT_CONFIG_KEY_0=url."https://${u}:${p}@github.com/".insteadof \
        GIT_CONFIG_VALUE_0=https://github.com/ \
        go mod download; \
    else \
        go mod download; \
    fi

# Copy the rest of the source code
COPY . .

# Build the application
RUN --mount=type=cache,target="/chorus/.cache/go-build" \
    --mount=type=cache,target="/chorus/.cache/go-mod" \
    go build -trimpath -ldflags "$LD_FLAGS" -o /usr/local/bin/chorus ./cmd/chorus

# Runtime stage - minimal image with only the binary and runtime dependencies
FROM harbor.build.chorus-tre.ch/docker_proxy/library/ubuntu:24.04

LABEL org.opencontainers.image.source="https://github.com/CHORUS-TRE/chorus-backend"
LABEL org.opencontainers.image.description="Chorus backend"
LABEL org.opencontainers.image.vendor="CHORUS-TRE"

# Create non-root user
RUN groupadd -g 10000 chorus && \
    useradd -u 10000 -g chorus -s /sbin/nologin -m chorus

# Install only runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create necessary directories for config and data
RUN mkdir -p /config /var/lib/chorus && \
    chown -R chorus:chorus /config /var/lib/chorus

# Copy the compiled binary from builder stage
COPY --from=builder /usr/local/bin/chorus /usr/local/bin/chorus

# Switch to non-root user
USER chorus

# Health check using the HTTP health endpoint
# Adjust the port if your application uses a different port
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=5 \
    CMD curl -f http://localhost:5000/v1/health || exit 1

ENTRYPOINT ["chorus"]
CMD ["start", "--config", "/config/config.yaml"]
