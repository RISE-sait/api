# Use the official Golang image as a base image
FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache coreutils

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

RUN for domain in event game haircut identity location membership program payment user; do \
        cd internal/domains/$domain/persistence/sqlc && \
        sqlc generate && \
        cd /app; \
    done

# Build the Go application
RUN go test ./...