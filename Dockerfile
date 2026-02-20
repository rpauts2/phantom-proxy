# PhantomProxy - Multi-stage production build
# Go 1.24 + minimal runtime

FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /phantom-proxy ./cmd/phantom-proxy

# Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /phantom-proxy .
COPY configs/ ./configs/
COPY config.yaml .

EXPOSE 443 8080

ENTRYPOINT ["./phantom-proxy"]
CMD ["--config", "config.yaml"]
