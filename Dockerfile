# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Generate go.sum
RUN go mod tidy

# Copy source code
COPY . .

# Build the application with verbose output
RUN echo "Building gateway..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o gateway ./cmd/gateway && \
    echo "Build complete. Binary details:" && \
    ls -la gateway

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary tools for debugging
RUN apk add --no-cache bash

# Copy the binary from builder
COPY --from=builder /app/gateway .

# Copy config files
COPY --from=builder /app/config ./config

# Copy and set up the run script
COPY run.sh .
RUN chmod 755 run.sh && \
    chmod 755 gateway && \
    echo "Final image contents:" && \
    ls -la

# Expose port
EXPOSE 8080

# Run the application using the shell script
CMD ["/bin/bash", "run.sh"] 