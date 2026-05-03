# CLAUDE.md — UB-Mager API

> Context file for AI assistants working on this codebase.

## Project Overview

Backend API for the **UB-Mager** ride-hailing platform (Universitas Brawijaya campus transportation). Handles authentication, ride lifecycle, real-time driver tracking, and admin operations.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 16 + PostGIS + TimescaleDB |
| Cache | Redis 7 (GeoSet, Pub/Sub, Caching) |
| WebSocket | gorilla/websocket |
| Routing Engine | OSRM (self-hosted) |
| Auth | JWT (golang-jwt/v5) |
| Logging | zerolog |

## Project Structure

```
cmd/
├── server/       → Main application entry point
├── migrate/      → Database migration commands (up, fresh)
└── seed/         → Database seeder

internal/
├── config/       → Environment configuration loading
├── database/     → PostgreSQL & Redis connection setup
├── model/        → Domain models (GORM structs)
├── repository/   → Data access layer (queries)
├── service/      → Business logic layer
├── handler/      → HTTP handlers (Gin route handlers)
├── middleware/   → Auth, CORS, rate limiting middleware
├── ws/           → WebSocket hub & client management
├── cache/        → Redis caching layer
└── pkg/          → Shared packages (OSRM client, geo utils, JWT helpers)
```

## Architecture Pattern

- **Layered architecture**: Handler → Service → Repository → Database
- Handlers parse requests and call services
- Services contain business logic and call repositories
- Repositories handle database queries via GORM
- Models are GORM structs with JSON tags for Gin serialization

## Commands

```bash
make dev              # Run dev server (go run ./cmd/server)
make build            # Build binary to bin/server
make test             # Run tests with race detection + coverage
make lint             # Run golangci-lint
make migrate-up       # Auto-migrate (create/alter tables)
make migrate-fresh    # Drop all tables and re-migrate
make migrate-refresh  # Fresh migration + seed
make seed             # Seed database with test data
make docker-all       # Start all infra (PostgreSQL, Redis, OSRM, pgAdmin)
make docker-db        # Start PostgreSQL + pgAdmin only
make docker-cache     # Start Redis only
make docker-routing   # Start OSRM only
make docker-down      # Stop all Docker services
make check-services   # Check if required services are running
```

## Infrastructure Services

| Service | Container | Port |
|---------|-----------|------|
| PostgreSQL + PostGIS + TimescaleDB | `ubmager-postgres` | 5432 |
| Redis | `ubmager-redis` | 6379 |
| pgAdmin | `ubmager-pgadmin` | 5050 |
| OSRM | `ubmager-osrm` | 5000 |

> Shared services (`postgres_server`, `wardayagate-redis`, `pgadmin_web`) may already be running from other wardayadev projects.

## Environment Variables

Copy `.env.example` to `.env`. Key variables:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — PostgreSQL connection
- `REDIS_HOST`, `REDIS_PORT` — Redis connection
- `JWT_SECRET` — JWT signing key
- `OSRM_URL` — OSRM routing engine URL
- `PORT` — Server port (default: 8081)

## Conventions

- **Module path**: `github.com/wardayadev/ub-mager-api`
- **Error handling**: Return errors up the call chain; handlers respond with appropriate HTTP status
- **Naming**: Go standard (camelCase for unexported, PascalCase for exported)
- **API prefix**: `/api/v1/`
- **Auth**: JWT Bearer token in Authorization header
- **WebSocket**: Used for real-time driver location tracking
- **Logging**: Use `zerolog` (structured JSON logging)

## Related Repos

- `ub-mager-ml` — ML demand prediction service (Python/FastAPI, port 8000)
- `ub-mager-dashboard` — Admin dashboard (Next.js, port 3000)
- `ub-mager-docs` — Bruno API collection
- `ub-mager-driver` — Driver mobile app (Flutter) — not actively developed
- `ub-mager-passenger` — Passenger mobile app (Flutter) — not actively developed
