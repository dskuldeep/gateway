# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Generate go.sum
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -o gateway .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache bash

# Copy the binary from builder
COPY --from=builder /app/gateway .

# Copy config files
COPY --from=builder /app/config ./config

# Copy and set up the run script
COPY run.sh .
RUN chmod +x run.sh gateway

# Expose port
EXPOSE 8080

# Run the application using the shell script
CMD ["/bin/bash", "run.sh"]