# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y%m%d-%H%M%S)" \
    -o analytics \
    cmd/analytics/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1000 analytics && \
    adduser -D -u 1000 -G analytics analytics

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/analytics .

# Copy any static files or templates if needed
# COPY --from=builder /app/templates ./templates

# Change ownership
RUN chown -R analytics:analytics /app

# Switch to non-root user
USER analytics

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Run the application
ENTRYPOINT ["./analytics"]