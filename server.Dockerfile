# Use the official Golang image as a base image
FROM golang:1.23.4-alpine AS builder

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Set the working directory inside the container
WORKDIR /app

# Install necessary dependencies
RUN apk add --no-cache coreutils

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

RUN cd internal/domains/course/persistence/sqlc && sqlc generate
RUN cd internal/domains/enrollment/persistence/sqlc && sqlc generate
RUN cd internal/domains/event/persistence/sqlc && sqlc generate
RUN cd internal/domains/event_staff/persistence/sqlc && sqlc generate
RUN cd internal/domains/game/persistence/sqlc && sqlc generate
RUN cd internal/domains/haircut/persistence/sqlc && sqlc generate
RUN cd internal/domains/identity/persistence/sqlc && sqlc generate
RUN cd internal/domains/location/persistence/sqlc && sqlc generate
RUN cd internal/domains/membership/persistence/sqlc && sqlc generate
RUN cd internal/domains/practice/persistence/sqlc && sqlc generate
RUN cd internal/domains/purchase/persistence/sqlc && sqlc generate
RUN cd internal/domains/user/persistence/sqlc && sqlc generate

# Build the Go application
RUN go build -o server cmd/server/main.go

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
