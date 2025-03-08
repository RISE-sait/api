# Build Stage
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Copy and install Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire app
COPY . .

# Run tests
ENTRYPOINT ["go", "test", "./..."]
