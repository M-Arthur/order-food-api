# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable Go modules
ENV GO111MODULE=on

# Install build deps
RUN apk add --no-cache git

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the server binary
RUN CGO_ENALBED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Expose HTTP port
EXPOSE 8080

# Environment defaults
ENV PORT=8080

ENTRYPOINT [ "/app/server" ]
