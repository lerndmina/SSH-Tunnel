# Multi-stage build for SSH Tunnel Manager
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make openssh-client

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    openssh-client \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1001 -S ssh-tunnel && \
    adduser -u 1001 -S ssh-tunnel -G ssh-tunnel

# Create necessary directories
RUN mkdir -p /home/ssh-tunnel/.ssh-tunnel-manager && \
    chown -R ssh-tunnel:ssh-tunnel /home/ssh-tunnel

# Copy binary from builder
COPY --from=builder /app/build/ssh-tunnel /usr/local/bin/ssh-tunnel

# Make binary executable
RUN chmod +x /usr/local/bin/ssh-tunnel

# Switch to non-root user
USER ssh-tunnel

# Set working directory
WORKDIR /home/ssh-tunnel

# Expose common SSH tunnel ports
EXPOSE 2222 1080

# Set default command
ENTRYPOINT ["ssh-tunnel"]
CMD ["--help"]

# Add labels
LABEL org.opencontainers.image.title="SSH Tunnel Manager"
LABEL org.opencontainers.image.description="Cross-platform SSH tunnel management tool"
LABEL org.opencontainers.image.source="https://github.com/yourusername/ssh-tunnel-manager"
LABEL org.opencontainers.image.licenses="MIT"
