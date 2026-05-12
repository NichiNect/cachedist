# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /cachedist-server ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /cachedist-server /cachedist-server

# Expose HTTP and gRPC ports
EXPOSE 7000 8000

# Run the binary
ENTRYPOINT ["/cachedist-server"]
