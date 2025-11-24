# Multi-stage Dockerfile for Hasura MCP Server
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire CLI source
COPY . ./

# Build the MCP server binary (only main.go, exclude mock-hasura.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o hasura-mcp-server \
    ./cmd/mcp-server/main.go

# Stage 2: Create minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 hasura && \
    adduser -D -u 1000 -G hasura hasura

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/hasura-mcp-server /app/hasura-mcp-server

# Copy LICENSE file (required by Apache 2.0)
COPY --from=builder /build/cmd/mcp-server/LICENSE /app/LICENSE

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Change ownership
RUN chown -R hasura:hasura /app

# Switch to non-root user
USER hasura

# Expose no ports (stdio-based MCP server)
# The MCP server communicates via stdio, not HTTP

# Set entrypoint
ENTRYPOINT ["/app/hasura-mcp-server"]

# Default command (can be overridden)
CMD ["--help"]

# Labels for Docker Hub
LABEL maintainer="Sanjay Ramadugu <sanjay3290>"
LABEL org.opencontainers.image.title="Hasura MCP Server"
LABEL org.opencontainers.image.description="Model Context Protocol server for Hasura GraphQL Engine - enables AI assistants to interact with Hasura"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="Sanjay Ramadugu"
LABEL org.opencontainers.image.url="https://github.com/hasura/graphql-engine"
LABEL org.opencontainers.image.documentation="https://github.com/hasura/graphql-engine/tree/master/cli/cmd/mcp-server"
LABEL org.opencontainers.image.source="https://github.com/hasura/graphql-engine"
LABEL org.opencontainers.image.licenses="Apache-2.0"
