# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates make protobuf

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/captcha cmd/proxy/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 -S captcha && \
    adduser -u 1000 -S captcha -G captcha

WORKDIR /app

COPY --from=builder /app/bin/captcha /app/captcha
COPY --from=builder /app/configs /app/configs

RUN chown -R captcha:captcha /app

USER captcha

EXPOSE 9001 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/live || exit 1

ENTRYPOINT ["/app/captcha", "--config", "/app/configs/config.yaml"]
