# core — shared infrastructure (singleton)

This directory is the **shared infrastructure** of the application. There is exactly one of it; it has no archetype because an archetype describes a category and `core/` is unique.

## Sub-packages

Each is a focused, framework-level utility used by 2+ modules:

- `app` — application-level configuration and JWT auth setup
- `contextx` — `ContextX` wrapper around `context.Context`; carries app config and authenticated user ID
- `cryptox` — encryption / decryption helpers
- `db` — SQLite connection management, migrations, transactions (`WithTx`)
- `httpx` — custom mux, middleware, error handling
- `sqlx` — SQL type helpers (e.g. nullable types)
- `task` — background task scheduling and execution; defines `Task` interface
- `templates` — shared Templ components (layouts, redirects, etc.)
- `timex` — time-related utilities and constants
- `utils` — small general-purpose helpers

For details on individual sub-packages see `README.md` in this directory.

## Rules for adding to core

- **Used by 2+ modules.** Single-consumer code stays in the consumer.
- **Framework-level, not domain.** No business concepts in core. If it mentions albums, ratings, tags, users — it doesn't belong here.
- **`x` suffix** marks extension packages over a stdlib counterpart (`timex`, `sqlx`, `cryptox`).

## Why core is not an archetype

Archetypes describe categories with multiple instances. There is one `core/`. Some sub-packages are stateless (`timex`, `sqlx`) and would fit the `utility` archetype's rules; others hold state or lifecycle responsibility (`db`, `task`) and don't. Domain modules and external clients are both allowed to import `core/*`, which contradicts utility's "no external client may import this" stance. Documenting core as a singleton avoids carving out exceptions to other archetypes' rules.
