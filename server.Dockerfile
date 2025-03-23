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

# Build the Go application
RUN go build -o server cmd/server/server/main.go

# Final lightweight image
FROM alpine:latest AS final

WORKDIR /root/

# Install only required dependencies
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port your Go server listens on
EXPOSE 80

# Run the compiled binary
CMD ["./server"]
