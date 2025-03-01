# Use the official Golang image for building the migration application
FROM golang:1.23.4-alpine AS builder

# Install necessary tools (including Goose for migrations)
RUN apk add --no-cache git curl
RUN curl -sSfL https://github.com/pressly/goose/releases/download/v3.24.1/goose_linux_arm64 -o /usr/local/bin/goose && \
    chmod +x /usr/local/bin/goose

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod .
COPY go.sum .

# Download Go dependencies
RUN go mod download

ENV GOOSE_DRIVER=postgres
ENV GOOSE_MIGRATION_DIR="./db/migrations"

# Copy everything else
COPY . .

ENTRYPOINT ["go", "run", "cmd/migrate/main.go"]