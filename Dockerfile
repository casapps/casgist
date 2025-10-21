# Multi-stage build for CasGists
# BASE SPEC compliant: Alpine, static binary, metadata labels
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# Builder stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build static binary (CGO_ENABLED=0 per BASE SPEC)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o casgists ./src/cmd/casgists

# Final stage - Alpine (not scratch) to include curl and bash per BASE SPEC
FROM alpine:latest

# Metadata labels per BASE SPEC
LABEL org.opencontainers.image.title="CasGists"
LABEL org.opencontainers.image.description="Production-ready self-hosted GitHub Gist alternative"
LABEL org.opencontainers.image.vendor="CasApps"
LABEL org.opencontainers.image.authors="CasApps <support@casapps.com>"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="https://github.com/casapps/casgists"
LABEL org.opencontainers.image.source="https://github.com/casapps/casgists"
LABEL org.opencontainers.image.documentation="https://github.com/casapps/casgists"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"

# Install runtime dependencies: curl and bash per BASE SPEC
RUN apk --no-cache add ca-certificates tzdata curl bash

# Create app user (uid/gid 1001 - in range 100-999 per BASE SPEC for system user)
RUN addgroup -g 1001 casgists && \
    adduser -D -s /bin/bash -u 1001 -G casgists casgists

# Create necessary directories per BASE SPEC
# DataDir: /data, ConfigDir: /config, LogDir: /var/log/casgists
RUN mkdir -p /data /data/db /config /var/log/casgists && \
    chown -R casgists:casgists /data /config /var/log/casgists

# Copy binary to /usr/local/bin per BASE SPEC
COPY --from=builder /app/casgists /usr/local/bin/casgists
RUN chmod +x /usr/local/bin/casgists

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Change to app user
USER casgists

# Set environment variables per BASE SPEC
ENV CASGISTS_DATA_DIR=/data
ENV CASGISTS_LOG_DIR=/var/log/casgists
ENV CASGISTS_DB_TYPE=sqlite
ENV CASGISTS_DB_DSN=/data/db/casgists.db
ENV CASGISTS_SERVER_PORT=80

# Expose port 80 (internal) per BASE SPEC
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:80/api/v1/health || exit 1

# Use entrypoint script
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/usr/local/bin/casgists"]
