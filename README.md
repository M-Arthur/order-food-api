# Order Food API (Go Backend)

This project implements a production-grade backend for the **Order Food Online** API challenge using Go, following clean architecture patterns and idiomatic Go practices.

The service exposes:
- **Product API**
- **Order API**
- **Promo code validation subsystem**
- **Graceful shutdown**
- **Structured logging**
- **Modular package layout**

---

## Project Structure

The repository follows a clean and idiomatic Go layout:

```
order-food-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── config/                  # Environment & configuration loading
│   ├── domain/                  # Pure domain models (Product, Order, etc.)
│   ├── http/                    # HTTP routes, handlers, middleware
│   ├── logger/                  # Structured logging (zerolog)
│   ├── promo/                   # Promo code validation + file loading
│   ├── server/                  # Server wrapper (start, shutdown)
│   ├── service/                 # Business logic: OrderService, ProductService
│   └── storage/                 # In-memory repositories or storage abstraction
├── hooks/                       # Git hooks (pre-commit)
├── Dockerfile                   # Container build
├── docker-compose.yml           # Local environment orchestration
├── Makefile                     # Developer workflow commands
├── go.mod
├── go.sum
└── README.md
```

### Directory Responsibilities

**cmd/server**  
- The entrypoint that initializes router, configuration, logging, and starts the HTTP server.

**internal/config**  
- Loads environment variables such as `PORT`, coupon file paths, API key, etc.

**internal/domain**  
- Contains plain Go structs that map directly to the OpenAPI definitions.

**internal/service**  
- Business logic for products, orders, and interactions between components.

**internal/storage**  
- Repository layer. The challenge uses an in-memory implementation.

**internal/promo**  
- Promo code validation subsystem.  
- Loads `couponbaseX.gz` files and validates promo codes across datasets.

**internal/http**  
- HTTP handlers, route registration, response helpers, and middleware.

**internal/server**  
- Wrapper around `http.Server` providing graceful shutdown support.

**internal/logger**  
- Centralized structured logging built with `zerolog`.

---

## Makefile Usage

The project includes a `Makefile` to simplify development tasks.

Available commands:

```
make run        # Start the API server (go run ./cmd/server)
make test       # Run all unit tests
make tidy       # Clean go.mod & go.sum
make build      # Build the server binary into /bin/server
make lint       # Run golangci-lint
```

### Typical Development Workflow

Start the server:

```bash
make run
```

Run the test suite:

```bash
make test
```

Run the linter:

```bash
make lint
```

Build a release binary:

```bash
make build
```

---

## Running the Project

### Run locally

```bash
PORT=8080 go run ./cmd/server
```

Or simply:

```bash
make run
```

API endpoints:

```
GET /product
GET /product/{id}
POST /order
GET /health
```

---

## Code Quality: golangci-lint

This project uses **golangci-lint** to enforce consistent, idiomatic, and safe Go code.  
Before committing any changes, please ensure that `golangci-lint` is installed on your system.

### Install golangci-lint (Ubuntu/Mac/Linux)

```bash
curl -sSfL https://raw.githubusercontent.com/golangci-lint/golangci-lint/master/install.sh | \
  sudo sh -s -- -b /usr/local/bin v2.6.2
```

Verify installation:

```bash
golangci-lint version
```

---

## Pre-commit Hook Setup

To automatically lint code before each commit, set up the Git `pre-commit` hook.

Copy the hook file:

```bash
cp hooks/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Now `golangci-lint` will automatically run whenever you commit changes.

---

## Summary

- Clean and extensible Go backend following best practices.
- Project structure designed for maintainability.
- `Makefile` simplifies running, testing, and building the project.
- `golangci-lint` enforces code quality.
- Optional Git pre-commit hook ensures linting before commits.
