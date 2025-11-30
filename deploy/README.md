# Order Food API – Deployment Guide (Docker + Postgres)

This README explains how to run the **Order Food API** and **Postgres database** using Docker Compose.

It also covers:

- Directory structure
- Database initialization using SQL migration files
- Environment configuration
- How to rebuild/reset the environment
- Useful development commands

---

## 1. Running with Docker Compose

### 1.1 Start the stack

From inside the `deploy/` directory:

```bash
cd deploy
docker compose up --build
```

This will:

- Start **Postgres** (`orderfood_db`)
- Build and start the **Order Food API** (`orderfood_api`)
- Mount SQL migrations into Postgres for automatic initialization
- Expose:
  - API at http://localhost:8080
  - Postgres at localhost:5432

### 1.2 Stopping the stack

```bash
docker compose down
```

### 1.3 Destroy the database volume (forces schema re-init)

If you modify SQL files in `db/migrations/`, Postgres **will NOT** re-run them unless you reset the volume.

Reset everything:

```bash
docker compose down -v
docker compose up --build
```

---

## 2. Environment Configuration

### 2.1 API environment variables

The API container uses:

```
DB_DSN=postgres://orderfood_user:orderfood_pass@db:5432/orderfood?sslmode=disable
PORT=8080
```

You may override these manually.

---

## 3. Database Initialization (Automatic)

All SQL files located under:

```
db/migrations/
```

are automatically executed by Postgres **on first startup** of the database volume.

Example file:

```
db/migrations/001_init.sql
```

Contents include creation of:

- `orders`
- `order_items`
- indexes

### Important:

**This folder is mounted in docker-compose:**

```yaml
- ../db/migrations:/docker-entrypoint-initdb.d:ro
```

And Postgres only runs these scripts the **first time the database is created**.

To force rerun:

```bash
docker compose down -v
docker compose up --build
```

---

## 4. Database Access

### Connect via CLI

```bash
docker exec -it orderfood_db psql -U orderfood_user -d orderfood
```

### Connect via GUI tools

Use:

```
Host: localhost
Port: 5432
User: orderfood_user
Password: orderfood_pass
Database: orderfood
```

---

## 6. Docker Compose File Reference

Located at:

```
deploy/docker-compose.yml
```

It defines:

- Postgres container  
- Auto-initialization from SQL  
- API container build  
- Environment variables  
- Volume configuration  

---

## 7. Dockerfile Reference

Located at project root:

```
Dockerfile
```

The image is built in two stages:

- Go build stage
- Distroless runtime stage

This produces a small, secure production-ready image.

---

## 8. Resetting Everything (Clean Rebuild)

```bash
cd deploy
docker compose down -v
docker compose up --build
```

This:

1. Removes containers
2. Removes the database volume
3. Re-runs SQL migrations
4. Rebuilds API binary
5. Starts fresh

---

## 9. Troubleshooting

### “relation does not exist” errors
The DB was created before SQL migrations were mounted.  
Fix:

```bash
docker compose down -v
docker compose up --build
```

### API cannot connect to DB
Check:

```bash
docker logs orderfood_db
docker logs orderfood_api
```

Ensure `DB_DSN` matches the credentials in compose.

### Changing SQL has no effect
Reset volume:

```bash
docker compose down -v
docker compose up --build
```

---
