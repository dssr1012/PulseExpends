# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pulse-expends ./cmd/mcp-server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/pulse-expends .

# Create non-root user
RUN addgroup -g 1000 pulse && \
    adduser -D -u 1000 -G pulse pulse && \
    chown -R pulse:pulse /root/pulse-expends

USER pulse

# Create necessary directories
RUN mkdir -p /home/pulse/.pulse-expends && \
    chown -R pulse:pulse /home/pulse/.pulse-expends

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./pulse-expends"]