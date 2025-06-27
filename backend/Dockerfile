# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o reconya-api ./cmd/main.go

# Runtime stage - with lightweight Firefox ESR for screenshots
FROM alpine:latest

# Install runtime dependencies with Firefox ESR (lighter than Chromium)
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    nmap \
    sudo \
    firefox-esr \
    xvfb \
    && rm -rf /var/cache/apk/* \
    && adduser -D -s /bin/sh reconya

# Set working directory
WORKDIR /app

# Copy binary from build stage and set permissions
COPY --from=builder /app/reconya-api /app/
RUN chmod +x /app/reconya-api

# Create directories with proper permissions
RUN mkdir -p /app/data \
    && chown -R reconya:reconya /app

# Switch to non-root user for security
USER reconya

# Expose port
EXPOSE 3008

# Command to run the application
CMD ["/app/reconya-api"]
