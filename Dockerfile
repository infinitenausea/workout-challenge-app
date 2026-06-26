# Build stage
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary (statically linked)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /workout-app main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy ca-certificates (needed for HTTPS requests like Telegram API)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary and frontend files from builder
COPY --from=builder /workout-app .
COPY --from=builder /app/frontend ./frontend

# Expose the app port
EXPOSE 8080

# Run the app
CMD ["./workout-app"]
