# Stage 1: Build the sandman binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X sandman/cmd.Version=latest" \
    -o sandman .

# Stage 2: Get Trivy
FROM ghcr.io/aquasecurity/trivy:latest AS trivy_engine

# Stage 3: Final image — based on ZAP (provides Java, Python, and all ZAP scan scripts)
FROM ghcr.io/zaproxy/zaproxy:stable

USER root

# Install curl (Opengrep) and ClamAV (malware scanning)
RUN apt-get update && apt-get install -y --no-install-recommends curl clamav \
    && rm -rf /var/lib/apt/lists/*

# Download the latest ClamAV virus definitions at build time
RUN freshclam

# Install Opengrep binary from official GitHub releases
RUN ARCH=$(dpkg --print-architecture) && \
    curl -sSL "https://github.com/opengrep/opengrep/releases/latest/download/opengrep-linux-${ARCH}" \
    -o /usr/local/bin/opengrep && \
    chmod +x /usr/local/bin/opengrep

# Copy Trivy from stage 2
COPY --from=trivy_engine /usr/local/bin/trivy /usr/local/bin/trivy

# Copy sandman binary from builder stage
COPY --from=builder /app/sandman /usr/local/bin/sandman

ENTRYPOINT ["sandman"]
