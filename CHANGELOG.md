# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-06

### Added

- **Auth**: Register, login, logout, refresh token with JWT
- **Auth**: Campus email validation (@student.ub.ac.id only)
- **Auth**: Password reset flow (request token + reset)
- **Auth**: Rate limiting on auth endpoints
- **Rides**: Full ride lifecycle (request → match → pickup → complete)
- **Rides**: Fare calculation with surge pricing (per vehicle type)
- **Rides**: Coordinate validation on all location inputs
- **Rides**: Campus zone tagging (pickup/dropoff)
- **Rides**: In-ride chat (WebSocket + HTTP persistence)
- **Tasks**: Task marketplace with 5 categories (Jastip Makanan, Barang, Titip Print, Antar Jemput, Other)
- **Tasks**: Task lifecycle (open → accepted → picking up → delivering → completed)
- **Tasks**: Task chat and task rating
- **Drivers**: Registration, verification, online/offline toggle
- **Drivers**: Real-time GPS location tracking via WebSocket
- **Drivers**: Earnings tracking (daily/weekly/monthly breakdown)
- **Drivers**: Acceptance rate calculation
- **Drivers**: Badge system (Verified Student, Verified Driver, Trusted Helper, Experienced, Top Rated)
- **Admin**: Dashboard stats with Redis caching (10s TTL)
- **Admin**: Driver management (list, verify, toggle online, detail)
- **Admin**: Ride management (list, detail, cancel, bulk cancel stuck)
- **Admin**: Task management (list with filters)
- **Admin**: User management (list, suspend, unsuspend, reset password)
- **Admin**: Report system (submit, list, resolve with admin notes)
- **Admin**: Audit log (automatic logging of all admin write actions)
- **Analytics**: Revenue stats, daily revenue, ride stats, peak hours, driver leaderboard
- **Notifications**: Per-user notification system (list, unread count, mark read)
- **Search**: Global search across rides, tasks, users (with pg_trgm indexes)
- **WebSocket**: Real-time driver location, ride status updates, chat messages, task notifications
- **WebSocket**: Auto-reconnection with exponential backoff
- **Security**: Input sanitization (XSS/HTML strip, SQL injection prevention)
- **Security**: Request body size limit (1MB)
- **Security**: CORS configuration
- **Security**: Secure refresh token cookies (httpOnly, SameSite)
- **Performance**: Gzip compression
- **Performance**: Redis caching for dashboard stats
- **Performance**: Composite database indexes for analytics queries
- **Performance**: pg_trgm trigram indexes for fast ILIKE search
- **Observability**: Structured JSON logging with zerolog
- **Observability**: Request correlation IDs (X-Request-ID)
- **Observability**: Metrics endpoint (/metrics)
- **Observability**: API versioning headers (X-API-Version)
- **Infrastructure**: Docker with health checks
- **Infrastructure**: docker-compose with PostgreSQL, Redis, OSRM
- **Infrastructure**: GitHub Actions CI (build + test + lint)
- **Infrastructure**: Database seeder with real UB campus data

### Technical Details

- Go 1.25, Gin framework
- PostgreSQL 16 + PostGIS + TimescaleDB
- Redis 7 (GeoSet, caching, driver status)
- gorilla/websocket for real-time communication
- OSRM for road-following route geometry
- 35 unit tests passing

[Unreleased]: https://github.com/wardayadev/ub-mager-api/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/wardayadev/ub-mager-api/releases/tag/v0.1.0
