# Build stage
FROM golang:1.22-alpine AS builder

# Install curl for health checks
RUN apk add --no-cache curl

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o reward-system ./cmd

# Run stage
FROM alpine:latest

# Install curl for health checks
RUN apk add --no-cache curl ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/reward-system .

# Copy environment file
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8080

# Run the application
CMD ["./reward-system"]