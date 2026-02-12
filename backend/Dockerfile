# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries - output to /app in builder
RUN go build -o /app/api cmd/api/main.go
RUN go build -o /app/worker cmd/worker/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates, tzdata, wget for healthcheck)
RUN apk add --no-cache ca-certificates tzdata wget

# Copy binaries from builder
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Create non-root user
RUN adduser -D orchestra
USER orchestra

# Default command (overridden in docker-compose)
CMD ["/app/api"]
