# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable Go modules
ENV GO111MODULE=on

# Install build deps
RUN apk add --no-cache git bash curl

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Install swag binary and ensure http-swagger is available
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go get github.com/swaggo/http-swagger@latest

# Generate swagger docs (docs will be created in ./docs)
RUN swag init -g cmd/server/main.go -o ./docs --parseDependency --parseInternal || true

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server
# Copy generated docs (if any)
COPY --from=builder /app/docs /app/docs

# Expose HTTP port
EXPOSE 8080

# Environment defaults
ENV PORT=8080

ENTRYPOINT [ "/app/server" ]
