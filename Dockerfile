# Multi-stage build for FDO Server Proxy
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the proxy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fdo-proxy ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S fdo && \
    adduser -u 1001 -S fdo -G fdo

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fdo-proxy .

# Create necessary directories
RUN mkdir -p /app/certs /app/logs /app/data && \
    chown -R fdo:fdo /app

# Switch to non-root user
USER fdo

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the proxy
ENTRYPOINT ["./fdo-proxy"] 