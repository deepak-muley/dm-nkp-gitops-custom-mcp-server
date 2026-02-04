# Dockerfile for dm-nkp-gitops-a2a-server
# Multi-stage build for minimal, secure production image

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the A2A server binary
# CGO_ENABLED=0 for static binary (required for distroless)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}" \
    -o /app/bin/dm-nkp-gitops-a2a-server \
    ./cmd/a2a-server

# =============================================================================
# Stage 2: Production Image
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

# Labels for container metadata
LABEL org.opencontainers.image.title="dm-nkp-gitops-a2a-server"
LABEL org.opencontainers.image.description="A2A (Agent-to-Agent) server for NKP GitOps infrastructure monitoring"
LABEL org.opencontainers.image.source="https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server"
LABEL org.opencontainers.image.vendor="Deepak Muley"
LABEL org.opencontainers.image.licenses="MIT"

# Copy the binary from builder
COPY --from=builder /app/bin/dm-nkp-gitops-a2a-server /dm-nkp-gitops-a2a-server

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the A2A server port
EXPOSE 8080

# Run as non-root user (65532 is the nonroot user in distroless)
USER 65532:65532

# Set the entrypoint
ENTRYPOINT ["/dm-nkp-gitops-a2a-server"]

# Default command arguments
CMD ["serve", "--read-only"]
