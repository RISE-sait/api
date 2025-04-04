# Use the official Golang image as a base image
FROM golang:1.23.4-alpine AS builder

RUN go install github.com/air-verse/air@latest && \
    apk add --no-cache coreutils && \
    rm -rf /var/cache/apk/*

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
EXPOSE 80

ENV DATABASE_URL=${DATABASE_URL}

ENTRYPOINT ["air"]
