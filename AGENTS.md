# Repository-local rules for Codex

Global rules still apply. Keep this file short and focused on this repository.

## Project Shape

- This repository is a Go CLI file indexer named `duf`.
- Main entrypoint: `cmd/go-indexer/main.go`.
- Core packages live under `internal/`:
  - `internal/scanner/` walks configured paths and builds file metadata.
  - `internal/database/` stores metadata and hashes in PostgreSQL.
  - `internal/hasher/` calculates SHA256 hashes.
  - `internal/config/` loads `config.yaml`.
- PostgreSQL schema lives in `migrations/PostgreSQL/0000.sql`; the app also creates the table through `database.InitSchema`.

## Local Runtime

- `config.yaml` and `docker-compose.yaml` are local runtime files and are ignored by Git.
- Use `config.yaml-sample` and `docker-compose.yaml-sample` as the tracked examples.
- PostgreSQL runs from Docker Compose as `duf_postgres`.
- The compose sample uses `postgres:18-alpine`, host port `55432`, and mounts the volume at `/var/lib/postgresql`.
- The sample DSN is `postgres://duf:duf@127.0.0.1:55432/duf?sslmode=disable`.
- Do not switch the PostgreSQL 18 volume mount back to `/var/lib/postgresql/data`.

## Commands

- Format Go code with:
  ```bash
  gofmt -w <files>
  ```
- Run tests with a sandbox-friendly cache:
  ```bash
  env GOCACHE=/private/tmp/duf-go-cache go test ./...
  ```
- Start the database with:
  ```bash
  docker compose up -d postgres
  ```
- Check database logs with:
  ```bash
  docker compose logs --tail=80 postgres
  ```

## Change Rules

- Prefer small, direct changes that match the existing package boundaries.
- Do not silently ignore database write errors; surface them clearly.
- Preserve the `json.Number` file ID flow unless changing the ID strategy intentionally.
- File IDs are unsigned 64-bit FNV values stored in PostgreSQL as `numeric(20, 0)` to avoid signed `bigint` overflow.
- Keep generated/local artifacts out of commits, including `.idea/`, `.DS_Store`, local configs, and backup data.
- Do not delete Docker volumes, backup data, or local config files unless the user explicitly asks.

## Verification

- For code changes, run `gofmt` and `env GOCACHE=/private/tmp/duf-go-cache go test ./...`.
- For Docker/PostgreSQL changes, also run `docker compose config` and verify `docker compose ps postgres` or logs.
