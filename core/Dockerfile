# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the server (API) and worker binaries
RUN go build -o /app/server cmd/server/main.go
RUN go build -o /app/worker cmd/worker/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /app/server /app/server
COPY --from=builder /app/worker /app/worker

RUN adduser -D orchestra
USER orchestra

CMD ["/app/server"]
