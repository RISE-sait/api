# Use the official Golang image as a base image
FROM golang:1.23.4-alpine AS builder

# Install necessary tools, including air
RUN apk add --no-cache git curl && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /usr/local/bin

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod .
COPY go.sum .

# Download and cache Go dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Expose the port your Go server listens on
EXPOSE 8080

# Default command for development using air
CMD ["go","run","cmd/server/main.go"]
