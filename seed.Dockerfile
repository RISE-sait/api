# Use the official Golang image for building the migration application
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

RUN go mod init api

RUN go get github.com/google/uuid && \
    go get github.com/lib/pq && \
    go get github.com/shopspring/decimal

# Copy everything else
COPY ./cmd/seed ./cmd/seed
COPY ./config ./config
COPY ./internal/custom_types ./internal/custom_types

ENTRYPOINT ["go", "run", "cmd/seed/main.go"]