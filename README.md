# UB-Mager API

> Backend API for the UB-Mager ride-hailing platform.

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL 16 + PostGIS 3.4 + TimescaleDB 2.26
- **Cache:** Redis 7 (GeoSet, Pub/Sub, Caching)
- **WebSocket:** gorilla/websocket
- **Routing:** OSRM (self-hosted)

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

## Shared Services (Already Running)

| Service | Container | Port | Image |
|---------|-----------|------|-------|
| PostgreSQL + PostGIS + TimescaleDB | `postgres_server` | `5432` | `postgis/postgis:16-3.4` |
| Redis | `wardayagate-redis` | `6379` | `redis:7-alpine` |
| pgAdmin | `pgadmin_web` | `5050` | `dpage/pgadmin4` |

> These are shared across wardayadev projects. No need to start them separately.

## Getting Started

```bash
# 1. Check shared services are running
make check-services

# 2. Start OSRM routing service (only extra service needed)
make osrm-up

# 3. Copy environment config
cp .env.example .env
# Edit .env with your postgres password

# 4. Run migrations
make migrate-up

# 5. Start development server (port 8081)
make dev
```

## API Documentation

See [ub-mager-docs](../ub-mager-docs) for Bruno API collections.

## Project Structure

```
cmd/server/          → Application entry point
internal/
├── config/          → Environment configuration
├── database/        → PostgreSQL & Redis connections
├── model/           → Domain models (GORM)
├── repository/      → Data access layer
├── service/         → Business logic
├── handler/         → HTTP handlers (Gin)
├── middleware/       → Auth, CORS, rate limiting
├── ws/              → WebSocket hub & clients
├── cache/           → Redis caching layer
├── worker/          → Background workers
└── pkg/             → Shared packages (OSRM, geo, JWT)
```
