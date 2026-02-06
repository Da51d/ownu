# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git curl

# Install Chalk for build metadata and SBOM generation
RUN CHALK_VERSION=$(curl -fsSL https://dl.crashoverride.run/chalk/current-version.txt) && \
    curl -fsSL -o /usr/local/bin/chalk "https://dl.crashoverride.run/chalk/chalk-${CHALK_VERSION}-Linux-$(uname -m)" && \
    chmod +x /usr/local/bin/chalk

# Copy source code
COPY . .

# Download dependencies and build
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ownu ./cmd/server

# Mark the binary with Chalk (adds build provenance metadata)
RUN chalk insert ownu

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binary from builder (with chalk mark embedded)
COPY --from=builder /app/ownu .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./ownu"]
