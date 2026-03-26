# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /axis \
    ./cmd/axis

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 axis

WORKDIR /home/axis

# Create data directory
RUN mkdir -p /var/lib/axis && chown -R axis:axis /var/lib/axis

# Copy binary from builder
COPY --from=builder /axis /usr/local/bin/

# Copy config if exists
COPY axis.yaml.example /etc/axis/axis.yaml.example

# Switch to non-root user
USER axis

# Default port
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/v1/health || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/axis"]
CMD ["serve"]
