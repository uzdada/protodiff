# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o protodiff \
    ./cmd/protodiff

# Final stage
FROM alpine:latest

# Install buf CLI and runtime dependencies
RUN apk add --no-cache ca-certificates && \
    wget -O /usr/local/bin/buf https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-$(uname -m) && \
    chmod +x /usr/local/bin/buf

# Copy the binary
COPY --from=builder /build/protodiff /protodiff

# Use non-root user
USER 65532:65532

# Expose web server port
EXPOSE 18080

# Run the application
ENTRYPOINT ["/protodiff"]
