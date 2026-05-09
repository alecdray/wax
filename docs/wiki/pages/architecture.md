---
description: >
  How the system is structured — modules, patterns, request lifecycle, and deployment model.
  Belongs here: module responsibilities, architectural patterns, data access strategy, background
  tasks, and infrastructure. Does not belong here: UI/rendering concerns (→ frontend), external
  service details (→ integrations), or domain entity definitions (→ data-model).
links:
  - data-model
  - integrations
  - frontend
  - testing
---

[Parent: wiki](../wiki.md)

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
| notes | Album sleeve notes |
| labels | Record labels (stub) |
| feed | Data sync from external sources |
| spotify | Spotify API client |
| musicbrainz | MusicBrainz metadata client |
| discogs | Discogs API client |
| listeninghistory | Play history tracking |
| genres | Genre DAG (utility) |

## Key Patterns

Architectural rules — module archetypes, allowed/forbidden imports, file layout, repo and adapter conventions — live in [`docs/architecture/`](../../architecture/). The wiki summarizes the system; the architecture docs are the authoritative source for rules.

## Deployment

Single binary + SQLite file, containerized with Docker. No external database or cache required. Static assets (CSS, JS) are bundled at build time.

