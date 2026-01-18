# Build stage
FROM golang:1.22-alpine AS builder

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

# Build the binary
# CGO_ENABLED=0 for static binary (required for scratch/distroless)
# -ldflags="-w -s" to strip debug info and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./main.go

# Final stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/server /server

# Copy migrations (if needed for runtime)
COPY --from=builder /app/migrations /migrations

# Use nonroot user (UID 65532)
USER nonroot:nonroot

# Expose port (Cloud Run will set PORT env var)
EXPOSE 8080

# Health check
# Note: distroless doesn't have curl/wget, so we rely on Cloud Run's health checks

# Run the server
ENTRYPOINT ["/server"]
