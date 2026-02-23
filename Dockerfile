FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o phantom-proxy ./cmd/phantom-proxy-v14

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/phantom-proxy /app/phantom-proxy
COPY --from=builder /app/config.yaml /app/config.yaml

# Create data directory
RUN mkdir -p /data

EXPOSE 8443 8080

CMD ["./phantom-proxy", "--config", "config.yaml"]