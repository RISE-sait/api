# Use the official Golang image as a base image
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download and cache Go dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o server ./cmd/server

# Start a new stage from a smaller base image
FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/.env .

# Expose the port your Go server listens on (change if needed)
EXPOSE 8080

# Command to run the Go server
CMD ["./server"]
