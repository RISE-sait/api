# Use the official Golang image for building the migration application
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

RUN go mod init api

RUN go get github.com/google/uuid github.com/lib/pq github.com/shopspring/decimal github.com/go-playground/validator

# Copy everything else
COPY ./cmd/seed ./cmd/seed
COPY ./config ./config
COPY ./internal/custom_types ./internal/custom_types
COPY ./internal/libs/validators ./internal/libs/validators
COPY ./internal/libs/errors ./internal/libs/errors

ENTRYPOINT ["go", "run", "cmd/seed/main.go"]