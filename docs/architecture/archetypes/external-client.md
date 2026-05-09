# External Client Module

## Purpose

An external client module wraps a third-party API. It has no own domain concepts, no persistence, and no HTTP entrypoints. It exists only to expose a clean, internal-facing interface over a remote service. Consumer domain modules depend on client modules — never the reverse.

## File layout

```
src/internal/<module>/
├── client.go       # Client struct + low-level HTTP/SDK calls — required
├── entities.go     # types from the external API; conversions to internal types — required
├── service.go      # internal-facing operations consumers use — required
├── *_test.go       # tests live next to the file under test
└── CLAUDE.md       # required; declares archetype
```

No `adapters/`. No `repo.go`. README is optional — a package doc comment in `client.go` is sufficient.

## Responsibilities by file

- **`client.go`** — owns the `Client` struct, authentication configuration, HTTP transport, and all raw API calls. Low-level: makes the request, decodes the response, returns entities from `entities.go` or stdlib types.
- **`entities.go`** — mirrors the external API's data shapes as Go types. Includes any conversion functions that translate external types to internal types consumed by the rest of the application. No business logic.
- **`service.go`** — owns the `Service` struct. Methods are the interface domain modules call. `Service` wraps `*Client`; it composes and filters raw client results into the shapes callers actually need.

## Allowed imports

- `core/*` sub-packages.
- Vendor SDKs for the wrapped API.
- Stdlib.

## Forbidden imports

- Any domain module. Client modules are leaves in the dependency graph: domain modules depend on clients, never the reverse.
- Other external client modules.
- `core/db/sqlc` — clients own no tables and must not touch the database.

## Why no DB / repo

External client modules do not own any database tables. If a third-party integration needs to persist data — cached API responses, OAuth tokens, rate-limit state — that persistence belongs in the consuming domain module, not in the client. The client fetches; the domain module decides what to store.

## Where new code goes

| Change | File |
|---|---|
| New raw API call | `client.go` |
| New external API type or conversion function | `entities.go` |
| New operation exposed to domain modules | `service.go` |
| Tests for any of the above | `*_test.go` next to the file under test |
