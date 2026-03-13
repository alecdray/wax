---
description: >
  How the system is structured — modules, patterns, request lifecycle, and deployment model.
  Belongs here: module responsibilities, architectural patterns, data access strategy, background
  tasks, and infrastructure. Does not belong here: UI/rendering concerns (→ frontend), external
  service details (→ integrations), or domain entity definitions (→ data-model).
links:
  - "[data-model](data-model.md)"
  - "[integrations](integrations.md)"
  - "[frontend](frontend.md)"
---

[wiki](../wiki.md)

# Architecture

How the system is structured at a high level.

## Stack

| Layer | Technology |
|---|---|
| Language | Go |
| Database | SQLite |
| Auth | JWT (cookies for web) |
| Migrations | Goose |
| SQL codegen | SQLC |
| Frontend | See [frontend](./frontend.md) |

## Module Structure

The application is organized into vertical modules, each owning a domain area end-to-end.

| Module | Responsibility |
|---|---|
| server | HTTP server, routing, service wiring |
| core | Shared infrastructure (context, HTTP utils, DB, tasks) |
| auth | Login flow, JWT issuance |
| user | User account management |
| library | Music collection (albums, artists, tracks, releases) |
| review | Album ratings and reviews |
| tags | User tagging system |
| feed | Data sync from external sources |
| spotify | Spotify API client |
| musicbrainz | MusicBrainz metadata client |
| listeninghistory | Play history tracking |

## Key Patterns

### Service Layer
Each module exposes a service that holds all business logic. Services are instantiated once at startup and injected where needed. There is no global state.

### Adapters
HTTP handlers and HTML templates are kept separate from business logic within each module. This separates the delivery mechanism (HTTP/HTML) from the domain.

### Request Flow

```
HTTP Request
  → Middleware (auth, logging)
    → Handler (adapters/)
      → Service (business logic)
        → Database (SQLC queries)
      ← DTO / domain model
    ← Templ component (HTML fragment)
  ← HTTP Response
```

### Context
A custom context type wraps the standard Go context and carries the authenticated user ID and app config throughout the request lifecycle.

### Error Handling
Errors are returned as HTML fragments for HTMX-driven pages, or JSON for API endpoints. A shared utility ensures consistent error responses across all handlers. See [frontend](./frontend.md) for the interaction model.

### Background Tasks
A task manager runs scheduled background jobs (e.g. Spotify library sync, listening history sync). Tasks implement a common interface with an ID, run function, and cron schedule.

### Database
- SQLite with connection pooling
- All queries are written in SQL and compiled to type-safe Go via SQLC — no ORM
- Transactions wrap multi-step operations
- Schema migrations run automatically on startup via Goose

## Deployment

Single binary + SQLite file, containerized with Docker. No external database or cache required. Static assets (CSS, JS) are bundled at build time.

