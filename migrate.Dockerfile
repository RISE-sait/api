# Use the official Golang image for building the migration application
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary tools (including Goose for migrations)
RUN go mod init api && \
    go get github.com/pressly/goose github.com/lib/pq github.com/stripe/stripe-go/v81

ENV GOOSE_DRIVER=postgres
ENV GOOSE_MIGRATION_DIR="./db/migrations"

# Copy everything else
COPY cmd/migrate ./cmd/migrate
COPY db ./db
COPY config ./config

ENTRYPOINT ["go", "run", "cmd/migrate/main.go"]