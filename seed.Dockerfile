# Use the official Golang image for building the migration application
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod .
COPY go.sum .

# Download Go dependencies
RUN go mod download

# Copy everything else
COPY . .

ENTRYPOINT ["go", "run", "cmd/seed/main.go"]