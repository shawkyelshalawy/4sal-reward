# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o reward-system ./cmd

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/reward-system .
COPY --from=builder /app/.env.example .env
EXPOSE 8080
CMD ["./reward-system"]