# Use official Go image as a builder
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Use minimal image for running the application
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy binary file from builder image
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run application
CMD ["./main"]