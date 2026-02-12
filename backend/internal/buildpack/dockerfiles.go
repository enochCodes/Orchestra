package buildpack

import "fmt"

// GenerateDockerfile returns a Dockerfile content based on framework and config
func GenerateDockerfile(framework, buildCmd, startCmd string) string {
	switch framework {
	case "node":
		return fmt.Sprintf(`FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN %s
CMD ["sh", "-c", "%s"]`, buildCmd, startCmd)

	case "go":
		return fmt.Sprintf(`FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN %s

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]`, buildCmd)

	case "python":
		return fmt.Sprintf(`FROM python:3.9-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
CMD ["sh", "-c", "%s"]`, startCmd)

	case "rust":
		return fmt.Sprintf(`FROM rust:1.75-alpine as builder
WORKDIR /usr/src/app
COPY . .
RUN %s

FROM alpine:3.18
COPY --from=builder /usr/src/app/target/release/app /usr/local/bin/app
CMD ["app"]`, buildCmd)

	case "nextjs-static":
		return fmt.Sprintf(`FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN %s

FROM nginx:alpine
COPY --from=builder /app/out /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]`, buildCmd)

	default:
		return "" // Assume repo has Dockerfile
	}
}
