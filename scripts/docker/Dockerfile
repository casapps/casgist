# Multi-stage Dockerfile for CasGists
# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -tags netgo -installsuffix netgo \
    -o casgists cmd/casgists/main.go

# Stage 2: Create the runtime image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    git \
    && update-ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh -u 1000 casgists

# Create required directories
RUN mkdir -p /data /app && \
    chown -R casgists:casgists /data /app

# Copy the binary from builder stage
COPY --from=builder /build/casgists /app/casgists
RUN chmod +x /app/casgists

# Copy static assets (if they exist)
# COPY --from=builder /build/web /app/web

# Switch to non-root user
USER casgists

# Set working directory
WORKDIR /app

# Expose default port
EXPOSE 64001

# Set environment variables
ENV CASGISTS_DATA_DIR=/data
ENV CASGISTS_LISTEN_PORT=64001
ENV CASGISTS_DB_PATH=/data/casgists.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/casgists --version || exit 1

# Default command
CMD ["/app/casgists"]