# Order Food API – Oolio Kart Backend Challenge

This repository contains a Go implementation of the **Oolio Kart “Order Food Online” backend challenge**.

It aims to look and behave like a small production service:

- HTTP API for **products**, **orders**, and **health**
- **PostgreSQL** persistence (orders/order items, products)
- **Promo code** processing pipeline + validation
- Clean, layered architecture (HTTP → service → storage → DB)
- **OpenAPI/Swagger** docs generated from code
- **Docker Compose** environment (API + Postgres) for easy review

The original challenge description is available at:
https://github.com/oolio-group/kart-challenge/blob/advanced-challenge/backend-challenge/README.md

---

## 1. Quickstart – How to Run

### 1.1 Run everything with Docker Compose (recommended)

Prerequisites:

- Docker
- Docker Compose (v2 or newer)

From the project root:

```bash
cd deploy
docker compose up --build
```

This will:

- Build the Go API image using the multi-stage `Dockerfile`
- Start **Postgres** (`orderfood_db`) with schema initialized from `db/migrations/`
- Start **Order Food API** (`orderfood_api`) and connect it to Postgres

Once up:

- API base URL: `http://localhost:8080`
- Health check: `GET http://localhost:8080/health`
- Product API: `GET http://localhost:8080/api/product`
- Order API: `POST http://localhost:8080/api/order`
- Swagger UI (if image was built with docs): `GET http://localhost:8080/swagger/`

To stop:

```bash
cd deploy
docker compose down
```

To reset DB (re-run migrations from scratch):

```bash
cd deploy
docker compose down -v
docker compose up --build
```

---

### 1.2 Run locally with Go

Prerequisites:

- Go (version compatible with `go.mod`, e.g. 1.21+)
- A running Postgres instance

1. Start/ensure Postgres is running.

   Example DSN (matches Docker Compose):

   ```bash
   export DB_DSN="postgres://orderfood_user:orderfood_pass@localhost:5432/orderfood?sslmode=disable"
   ```

2. Set environment variables and run the server:

   ```bash
   export PORT=8080
   # optional: export API_KEY=apitest
   go run ./cmd/server
   ```

   Or using the `Makefile`:

   ```bash
   make run
   ```

3. Hit the endpoints:

   ```text
   GET  /health
   GET  /api/product
   GET  /api/product/{productId}
   POST /api/order    (requires JSON body)
   ```

---

## 2. Project Structure & Architecture

High-level layout:

```text
cmd/
  server/            # HTTP API entrypoint (main.go)
  promo-loader/      # CLI tool to preprocess promo codes

internal/
  api/               # DTOs that mirror OpenAPI schemas + mappers
  bootstrap/         # Dependency wiring (services, repos, handlers)
  config/            # Env + config loading
  domain/            # Pure domain models & errors
  httpapi/           # Router, handlers, middleware, shared responses
  logger/            # Zerolog-based structured logger
  server/            # HTTP server wrapper (start/shutdown)
  service/           # Business logic (OrderService, ProductService)
  storage/           # Postgres repositories (orders, products)

db/migrations/       # SQL migrations to init Postgres schema

deploy/docker-compose.yml  # Local stack: Postgres + API
Dockerfile                  # Multi-stage build for API image
Makefile                    # Dev helpers (run, test, lint, docs)
valid_promo_codes.txt       # Output of promo-loader (used for promo code validation)
```

**Flow (simplified):**

```text
HTTP request → chi router (internal/httpapi/router.go)
            → handler (internal/httpapi/handlers/*)
            → service (internal/service/*)
            → repository (internal/storage/*)
            → Postgres (db/migrations schema)
```

Cross-cutting concerns:

- **Logging** – `internal/logger`, plus request logging middleware
- **Recovery** – panic-safe middleware returning JSON errors
- **JSON-only API** – content-type enforcement middleware
- **Graceful shutdown** – signal handling + `http.Server.Shutdown` in `cmd/server/main.go`

---

## 3. Features & Endpoints

### 3.1 Product API

Implemented in `internal/httpapi/handlers/product_handler.go` and `internal/service/product_service.go`.

- `GET /product`
  - Lists all products available for ordering.
- `GET /product/{productId}`
  - Returns a single product by ID (path param).
  - Validates that `productId` is a numeric ID.

### 3.2 Order API

Implemented in `internal/httpapi/handlers/order_handler.go` and `internal/service/order_service.go`.

- `POST /order`
  - Request body: `OrderReqDTO` (`internal/api/dto.go`), e.g.:
    ```json
    {
      "couponCode": "PROMO123",
      "items": [
        { "productId": "10", "quantity": 2 }
      ]
    }
    ```
  - Validates:
    - JSON shape and required fields
    - Product IDs are known
    - Quantities are positive
  - Applies promo validation (if `couponCode` is provided and present in `valid_promo_codes.txt`).
  - Response body: `OrderDTO` with order ID, items, resolved products.

Protected by API key middleware (see 3.4).

### 3.3 Health

- `GET /health`
  - Simple health endpoint implemented in `internal/httpapi/handlers/health_handler.go`.

### 3.4 Authentication / API Key

- The **order** endpoint requires an API key header.
- Security scheme matches the challenge’s OpenAPI description:
  - Header name: `api_key`
  - Example: `api_key: apitest`
- Implemented in `internal/httpapi/middleware/auth_api_key.go`.

---

## 4. OpenAPI / Swagger

OpenAPI docs are generated from Go code using **swag** (`github.com/swaggo/swag`).

Generated artifacts live under `./docs` (created during Docker build or via `make docs`):

- `docs/docs.go` – Go bindings for Swagger
- `docs/swagger.json` – JSON spec
- `docs/swagger.yaml` – YAML spec

### 4.1 Viewing the API Docs

If you run via Docker Compose (with the current `Dockerfile`):

- Swagger JSON: `GET http://localhost:8080/swagger/doc.json`
- Swagger UI: `GET http://localhost:8080/swagger/`

> The router serves `./docs/swagger.json` via `/swagger/doc.json`, and the Swagger UI handler points to that URL.

You can also import `docs/swagger.yaml` into your preferred API client (e.g. Swagger Editor, Postman) when working locally.

### 4.2 Regenerating Docs (optional)

Only needed if you change handlers/DTOs.

Prerequisite: install `swag` CLI:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

From repo root:

```bash
make docs
# or directly
swag init -g cmd/server/main.go -o ./docs --parseDependency --parseInternal
```

---

## 5. Promo Loader CLI (`cmd/promo-loader`)

The challenge requires dealing with **very large gzip-compressed promo code files**, keeping codes
that:

- are between **8 and 10 characters** long, and
- appear in **at least 2 different source files**.

This is implemented as a separate CLI tool: `cmd/promo-loader`.

### 5.1 What it does

- Streams `.gz` files line-by-line (no full file in memory).
- Buckets promo codes using hashing to limit memory usage.
- Counts appearances across input files.
- Writes all codes that meet the criteria into `valid_promo_codes.txt`.

The API loads this file at startup and uses it to validate the `couponCode` in orders.

### 5.2 How to run it

From repo root:

```bash
# Example (adjust paths/flags as needed)
 go run ./cmd/promo-loader \
  --files=tmp/promo/file_1.gz,tmp/promo/file_2.gz,tmp/promo/file_3.gz \
  --tmp-dir=./tmp/promo \
  --output=./valid_promo_codes.txt \
  --parallelism=3
```

See `cmd/promo-loader/README.md` for detailed flags and examples.

After running the loader, restart the API so it picks up the updated `valid_promo_codes.txt`.

---

## 6. Development Workflow

### 6.1 Makefile Commands

From repo root:

```bash
make run        # Start API (go run ./cmd/server) – expects DB_DSN to be set
make test       # Run all unit tests
make tidy       # go mod tidy
make build      # Build Docker images via docker compose (see deploy/Makefile)
make lint       # Run golangci-lint (if installed)
make docs       # Generate OpenAPI docs via swag into ./docs
```

### 6.2 Testing

Run the full test suite:

```bash
go test ./...
# or
make test
```

Coverage includes:

- Domain models: `internal/domain/models_test.go`
- Services: `internal/service/*_test.go`
- HTTP handlers: `internal/httpapi/handlers/*_test.go`
- Middleware: `internal/httpapi/middleware/*_test.go`
- Server wrapper: `internal/server/server_test.go`

### 6.3 Linting & Pre-Commit

Install `golangci-lint` (example):

```bash
curl -sSfL https://raw.githubusercontent.com/golangci-lint/golangci-lint/master/install.sh \
  | sudo sh -s -- -b /usr/local/bin v1.60.0
```

Run manually:

```bash
make lint
```

Enable pre-commit hook:

```bash
cp hooks/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

This will run linters automatically on each commit.

---

## 7. Mapping to the Oolio Kart Challenge

This section briefly connects the implementation back to the challenge requirements.

### 7.1 Product Catalog

- **Requirement:** list and fetch products.
- **Implementation:**
  - `GET /product`, `GET /product/{productId}`
  - Code: `internal/httpapi/handlers/product_handler.go`, `internal/service/product_service.go`, `internal/storage/postgres_product_repository.go`
  - Data: Postgres schema defined in `db/migrations/001_init.sql`.

### 7.2 Orders & Validation

- **Requirement:** place orders with validation.
- **Implementation:**
  - `POST /order` using `OrderReqDTO` / `OrderDTO` (`internal/api/dto.go`).
  - Validates input shape, product existence, quantities, and (optionally) promo code.
  - Returns appropriate HTTP errors with JSON body `shared.ErrorResponse`.

### 7.3 Promo Code Handling

- **Requirement:** handle very large promo datasets and apply them to orders.
- **Implementation:**
  - `cmd/promo-loader` for scalable preprocessing of `.gz` files.
  - Output `valid_promo_codes.txt` is used by the API at startup.
  - Orders with `couponCode` are checked against this set.

### 7.4 Persistence & DB

- **Requirement:** persist data.
- **Implementation:**
  - Postgres-backed repositories in `internal/storage/`.
  - Schema in `db/migrations/001_init.sql`.
  - Docker Compose brings up a ready-to-use DB for reviewers.

### 7.5 API Documentation

- **Requirement:** clear machine-readable API contract.
- **Implementation:**
  - OpenAPI 3.x spec generated via `swag` into `./docs`.
  - Swagger UI available at `/swagger/` when running via Docker.

---

## 8. Suggested Reading Order for Reviewers

If you’re reviewing this as a challenge submission, a good reading path is:

1. `cmd/server/main.go` – entrypoint, wiring, graceful shutdown
2. `internal/httpapi/router.go` – routes, middleware, Swagger annotations
3. `internal/httpapi/handlers/order_handler.go` & `product_handler.go` – HTTP layer
4. `internal/service/order_service.go` & `product_service.go` – business logic
5. `internal/storage/postgres_*.go` – persistence strategy
6. `db/migrations/001_init.sql` – DB schema
7. `cmd/promo-loader` – large promo file processing pipeline
8. `docs/swagger.yaml` – generated API contract

This should give a clear picture of how the codebase addresses the original Oolio Kart backend challenge.
