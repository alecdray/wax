# Wax Architecture Patterns

## Goals & Non-Goals

### Goals
- Lock in a small fixed set of **module archetypes** for `src/internal/`, each with concrete file and import rules an agent can follow without judgment.
- Encode the rules so agents auto-load them via layered `CLAUDE.md` files (root + per-module), backed by detailed archetype docs in a new `docs/architecture/` tree.
- Produce, as an output of this project, a **refactor backlog** that classifies every existing module and lists the changes — both compliance gaps and structural reorgs — needed to make it conform.

### Non-Goals
- Executing the refactor. That is downstream work driven by the backlog this project produces.
- Adding hooks or other automated enforcement. May be revisited if drift continues despite the documented rules.
- Changing the tech stack, build system, or wiki structure beyond pointing the existing wiki architecture page at the new docs.
- No changes to `src/internal/<module>/*.go`. No new repos, no moves, no extractions.

## Background

Today the wiki at `docs/wiki/pages/architecture.md` describes the architecture conceptually but is failing as guidance for agents because:
- Agents don't read it by default — it's not in their auto-loaded path.
- It's drifted out of sync with the code.
- It's too high-level — it describes intent, not concrete rules.

Modules under `src/internal/` have visibly drifted. Some follow the documented "service + adapters" shape (`library`, `review`, `tags`, `notes`, `labels`); others diverge in different directions (`auth` has only `http.go` and templ components; `spotify`, `musicbrainz`, `discogs` are flat client wrappers; `genres` is pure data; `feed` and `listeninghistory` have services with no `adapters/`). The wiki's module table also omits `discogs` entirely.

Underneath this, there is no repository layer. Services import `core/db/sqlc` directly, and SQLC's generated `Queries` struct exposes every query in the application from a single type, so any service can reach any table. SQLC types leak into service signatures (e.g. `library/service.go` helpers accept `sqlc.Release`, `sqlc.UserRelease`).

`library/service.go` is 1009 lines and spans albums, artists, tracks, releases, formats, and user-collection state — a candidate for structural reorg, not just compliance.

`server/server.go` (195 lines today) is the other obvious hot spot. It owns service construction, task registration, lifecycle, middleware setup, **and route registration for every module**. Every new feature touches it. URL patterns for `library`, `tags`, `notes`, `review`, and `auth` are all declared there — far from the handlers that serve them.

## Module Archetypes

Most modules under `src/internal/` are classified into exactly one of three archetypes. Two directories are **singletons** — there is only ever one of each — and are documented separately (see *Singletons* below).

### 1. Domain module

Owns a slice of the application's domain end-to-end: business logic, persistence, and (optionally) HTTP entrypoints.

**Files**
```
src/internal/<module>/
├── service.go          # Service struct + business logic; required
├── repo.go             # ONLY file that imports sqlc; required
├── <topic>.go          # pure-logic types/functions, split by topic
├── *_test.go           # tests live next to the file under test
├── README.md           # required
├── CLAUDE.md           # required; declares archetype
└── adapters/           # required if module has HTTP entrypoints
    ├── http.go         # HttpHandler struct + handler methods
    ├── routes.go       # RegisterRoutes(mux *httpx.Mux, h *HttpHandler) — URL patterns
    └── *.templ / *_templ.go
```

**Rules**
- `service.go` holds the `Service` struct. All business logic lives on the service.
- `repo.go` is the **only** file in the module allowed to import `core/db/sqlc`. It exposes domain-named methods that return DTOs / domain types — never `sqlc.*` types.
- Repo is a concrete struct. Services hold `*Repo`. If a service test needs a mock, the service defines a small interface locally (Go convention: accept interfaces, return structs).
- For multi-step transactions: `core/db.WithTx(ctx, func(tx) { ... })` at the service layer. Repo constructors take a `*sqlc.Queries` so they can be bound either to the global handle or to a transaction.
- Domain types (DTOs, value objects) live in module-root files named by topic (e.g. `review/rating.go`, `review/state.go`). Avoid a catch-all `models.go`.

**Allowed imports**
- `core/*`
- Other domain modules — but **only** their exported `Service` and DTO types, injected via constructor.

**Forbidden imports**
- Other domain modules' `adapters/`.
- Other domain modules' internal (non-exported) packages or types.
- `sqlc` from any file other than `repo.go`.

**Adapters**
- HTTP handlers in `adapters/http.go`. Templ files in `adapters/*.templ`, generated to `*_templ.go`.
- URL patterns and route registration live in `adapters/routes.go`. The function signature is `RegisterRoutes(mux *httpx.Mux, h *HttpHandler)`. Server passes the appropriate mux (root for public routes, app submux for authenticated routes) — the module decides which paths and methods bind to which handler methods.
- Adapters import their own module's `*Service` only. They do **not** import `repo.go`, `sqlc`, or peer modules.
- A handler that needs another module's data takes the peer `*Service` through its constructor. The wiring happens in `server/`.

**Module sizing & splits**

Line count is a trigger to look, not a reason to split. The actual question is whether the module owns one coherent responsibility.

*Indicators worth investigating* (any one prompts a closer look):
- `service.go` materially larger than peers.
- Frequent edits to the module touch only one of its sub-areas (high churn asymmetry).
- A new feature in this area routinely changes only a subset of the module.
- The module is imported from another module mostly for the needs of one specific sub-area.

*Justification to split* (must hold to actually recommend a split):
- The module owns multiple distinct domain concepts that don't share state.
- One sub-area would be reusable from a module that doesn't currently use this one.
- The sub-areas can be described independently without referring to each other.

*Don't split when*
- The two halves would constantly call each other — that is one aggregate, not two modules.
- One side is just data the other side owns — that is an internal type, not a module.

### 2. External client module

Wraps a third-party API. Has no own domain concepts and no HTTP entrypoints.

**Files**
```
src/internal/<module>/
├── client.go           # SDK / HTTP client + low-level calls
├── entities.go         # types from the external API and conversion to internal types
├── service.go          # internal-facing operations consumers use
├── *_test.go
└── CLAUDE.md
```

No `adapters/`. No `repo.go`. README is optional — a doc comment on the package in `client.go` is sufficient.

**Allowed imports**
- `core/*`
- Vendor SDKs and stdlib.

**Forbidden imports**
- Domain modules. Clients are leaves; domain modules depend on clients, never the reverse.

### 3. Utility module

Stateless: pure functions and/or embedded data. No service struct, no DB, no state.

**Files**
```
src/internal/<module>/
├── <name>.go
├── *_test.go
├── data.json (or similar)   # optional embedded data
└── CLAUDE.md
```

**Allowed imports**
- `core/*`, stdlib, small focused libraries.

**Forbidden imports**
- Domain modules.
- External clients.

## Singletons

Two directories under `src/internal/` exist as exactly-one-of-them. Their rules are documented in their own `CLAUDE.md` rather than as archetype docs, because an "archetype" describes a category and a category-of-one is just ceremony.

### `server/` — composition root

Builds services, sets up middleware and sub-muxes, calls each domain module's `RegisterRoutes`, and runs lifecycle (DB open, task manager, HTTP listener). Has no domain logic. Has no URL patterns of its own beyond mounting sub-muxes.

`src/internal/server/CLAUDE.md` documents these rules directly. Trying to fit server into an existing archetype would require carving out exceptions to that archetype's import rules — server is the only place allowed to import every domain module and external client, which is the opposite of what utility allows.

### `core/` — shared infrastructure

The `src/internal/core/` tree contains focused, framework-level sub-packages used by multiple modules: `contextx`, `httpx`, `db`, `task`, `templates`, `app`, `cryptox`, `sqlx`, `timex`, `utils`. Some are stateless (`timex`, `sqlx`), some hold state or lifecycle responsibility (`db`, `task`).

`src/internal/core/CLAUDE.md` documents the rules for what belongs in core:

- Used by **2+ modules** — single-consumer code stays in the consumer.
- **Framework-level**, not domain — no business concepts in core.
- The `x` suffix marks extension packages over a stdlib counterpart (`timex`, `sqlx`, `cryptox`).

Core sub-packages don't fit cleanly under "utility" because they are allowed to be imported by every other module (utility is forbidden from being imported by external clients) and some of them carry state.

## Module Classification

Every existing module is classified into one archetype. This classification is declared in the per-module `CLAUDE.md`, which is the single source of truth (no separate registry — see *Encoding* below).

The classification proposed by this project (subject to confirmation during execution):

| Module | Archetype | Notes |
|---|---|---|
| `auth` | domain | Currently incomplete — see backlog. |
| `core` | — (singleton) | Shared infrastructure. No archetype. Rules live in `src/internal/core/CLAUDE.md`. |
| `discogs` | external-client | |
| `feed` | domain | No HTTP, so no `adapters/`. |
| `genres` | utility | |
| `labels` | domain | Currently empty `adapters/`, no service — see backlog. |
| `library` | domain | Structural reorg candidate — see backlog. |
| `listeninghistory` | domain | |
| `musicbrainz` | external-client | |
| `notes` | domain | |
| `review` | domain | Closest to the documented shape. |
| `server` | — (singleton) | Composition root. No archetype. Rules live in `src/internal/server/CLAUDE.md`. |
| `spotify` | external-client | |
| `tags` | domain | |
| `user` | domain | |

`server/` and `core/` are documented above under *Singletons*. They each get a `CLAUDE.md` that captures their rules directly, and `docs/architecture/README.md` flags both so agents reading the archetype catalog know where their rules live.

## Encoding (file layout)

All paths relative to `projects/wax/`.

### New files

```
docs/architecture/
├── README.md                                  # index, archetype catalog, singletons note, navigation
├── archetypes/
│   ├── domain-module.md
│   ├── external-client.md
│   └── utility.md
└── refactor-backlog.md                        # per-module gap list
```

Singletons (`server/`, `core/`) are documented in their own `src/internal/<name>/CLAUDE.md` files, not in the `archetypes/` directory.

### Updated files

- `.claude/CLAUDE.md` — replace the "Architecture Patterns" section. New version is short: archetypes exist; every module declares its archetype in its own `CLAUDE.md`; full rules in `docs/architecture/`. Build/test/HTMX/database rules already in this file are preserved.
- `docs/wiki/pages/architecture.md` — keep the stack table and high-level module list. Replace the "Key Patterns" section with a one-line pointer to `docs/architecture/`.

### Per-module CLAUDE.md

One per module under `src/internal/`. ~10 lines, identical structure:

```
# <module> — <archetype>

Rules: ../../../docs/architecture/archetypes/<archetype>.md

Module-specific notes:
- <anything non-obvious about this module>
```

Why per-module instead of one big root file:
- Auto-loaded by Claude Code only when working in that subtree, keeping context focused.
- The archetype declaration sits next to the code, so renaming or moving the module doesn't orphan the rules.
- Makes the archetype an explicit, grep-able fact rather than implicit from file layout.

### Why no module registry

A separate `module-registry.md` would duplicate what each per-module `CLAUDE.md` already declares — the same drift trap the wiki already fell into. The at-a-glance view is recovered with a one-liner included in `docs/architecture/README.md`:

```
grep -h "^# " src/internal/*/CLAUDE.md
```

`refactor-backlog.md` survives because it is a TODO list, not a description of state — it does not drift the same way.

## Refactor Backlog

Generated as part of this project, in `docs/architecture/refactor-backlog.md`. Each existing module gets an entry with two parts:

1. **Compliance gaps** — missing files, files in the wrong place, forbidden imports, repo not extracted, routes still declared in `server` instead of in the module's `adapters/routes.go`.
2. **Structural notes** — recommended splits, merges, moves under `core`, or renames, with the responsibility-coherence rationale that justifies them.

Backlog entries are 1–3 sentences. The detailed design of any actual split is left to the refactor work that consumes this backlog.

A heuristic pass — line counts, a quick read of `service.go`, looking for responsibility incoherence — is sufficient. No deep file-by-file mapping at this stage.

`server/` gets its own backlog entry alongside the modules: extract per-module route registration into each domain module's `adapters/routes.go`, leaving server with services construction, middleware/sub-mux setup, lifecycle, and `RegisterRoutes` calls.

Anchoring example: the entry for `library` says roughly *"Compliance: extract `repo.go`; SQLC import disappears from `service.go`. Structural: library currently owns albums, artists, tracks, releases, formats, and user-collection state. Artists and tracks are independently meaningful concepts that other modules could plausibly reference directly; propose extracting `artists` and `tracks` as separate domain modules with `library` keeping the user-collection aggregate."*

## Definition of Done

This project is complete when **all** of the following are true:

1. Archetype docs exist at `docs/architecture/archetypes/{domain-module,external-client,utility}.md`. Each is concrete enough that an agent can answer, without judgment:
   - What files belong in this archetype?
   - What imports are allowed and forbidden?
   - Where does new code (a query, a handler, a background task) go?
2. `docs/architecture/README.md` indexes the archetypes, explains the layered-CLAUDE.md mechanism (including the grep one-liner), and flags `server/` and `core/` as singletons with pointers to their respective `CLAUDE.md` files.
3. Every directory under `src/internal/` has a `CLAUDE.md`. Domain-module / external-client / utility directories declare their archetype; `server/` and `core/` document their singleton rules directly. 15 files total, one per directory in `src/internal/`.
4. Root `.claude/CLAUDE.md` has its "Architecture Patterns" section replaced with a short pointer to `docs/architecture/`. The build/test/HTMX/database rules already in that file are preserved.
5. `docs/wiki/pages/architecture.md` has its "Key Patterns" section replaced with a one-line pointer to `docs/architecture/`. The stack table and module list are preserved.
6. `docs/architecture/refactor-backlog.md` lists, per non-compliant module, the specific compliance gaps and structural notes needed to comply.
7. Sanity check: a fresh agent given only `.claude/CLAUDE.md` and a task like *"add a new query to the library module"* produces code that lands in the right files with the right imports, without further guidance.

## Out of Scope

- Code changes in `src/internal/<module>/*.go` (no refactor, no new repos, no moves, no extractions).
- Hooks or `settings.json` changes.
- Wiki restructuring beyond the architecture-page pointer.
- The actual decision of whether and how to split `library` — captured as a backlog entry with rationale, not executed.
