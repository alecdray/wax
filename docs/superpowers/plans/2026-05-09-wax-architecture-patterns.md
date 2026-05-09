# Wax Architecture Patterns Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define and encode three module archetypes (plus rules for two singleton directories) for `src/internal/`, so that agents working in the wax codebase auto-load concrete, enforceable architectural rules instead of relying on a high-level wiki page they don't read.

**Architecture:** Layered `CLAUDE.md` mechanism — root `.claude/CLAUDE.md` points at `docs/architecture/`, which holds three archetype docs plus a README. Each directory under `src/internal/` gets its own `CLAUDE.md` declaring archetype (or, for `server`/`core`, documenting singleton rules). Refactor backlog produced as an output, not executed.

**Tech Stack:** Markdown only. No code changes in `src/internal/<module>/*.go`.

**Reference spec:** `docs/superpowers/specs/2026-05-09-wax-architecture-patterns-design.md`

---

## File Structure

**New files (5 in `docs/architecture/`)**
- `docs/architecture/README.md` — index, archetype catalog, singletons note, grep one-liner
- `docs/architecture/archetypes/domain-module.md` — full rules for the most common archetype
- `docs/architecture/archetypes/external-client.md` — rules for 3rd-party API wrappers
- `docs/architecture/archetypes/utility.md` — rules for stateless modules
- `docs/architecture/refactor-backlog.md` — per-module compliance + structural notes

**New files (15 `CLAUDE.md` files in `src/internal/`)** — one per existing directory:
`auth, core, discogs, feed, genres, labels, library, listeninghistory, musicbrainz, notes, review, server, spotify, tags, user`

**Modified files (2)**
- `.claude/CLAUDE.md` — replace "Architecture Patterns" section with pointer to `docs/architecture/`
- `docs/wiki/pages/architecture.md` — trim "Key Patterns" section to a one-line pointer

**Untouched** — no `.go` files. No `taskfile.yml`. No `settings.json`. No code in `src/`.

---

## Task 1: Write `archetypes/domain-module.md`

**Files**
- Create: `docs/architecture/archetypes/domain-module.md`

This is the most important and most detailed archetype doc — most modules in wax are domain modules. The doc is the authoritative reference that per-module `CLAUDE.md` files will link to.

- [ ] **Step 1: Read the spec section that defines this archetype**

Read `docs/superpowers/specs/2026-05-09-wax-architecture-patterns-design.md` lines covering the `Domain module` section (file layout, rules, allowed/forbidden imports, adapters with routes.go, module sizing & splits).

- [ ] **Step 2: Read two existing modules to ground the rules in real code**

Read `src/internal/review/` (closest to the documented shape) and `src/internal/library/service.go` (largest, will inform "module sizing & splits" examples). Note the actual import patterns and file shapes.

- [ ] **Step 3: Write `docs/architecture/archetypes/domain-module.md`**

Content must include, at minimum:

1. **Purpose** — one paragraph: a domain module owns a slice of business logic + persistence + (optionally) HTTP entrypoints, end to end.
2. **File layout** — show the canonical tree (`service.go`, `repo.go`, `<topic>.go`, `*_test.go`, `README.md`, `CLAUDE.md`, `adapters/{http.go, routes.go, *.templ, *_templ.go}`). State which files are required vs. optional. State that `adapters/` is required only if the module has HTTP entrypoints.
3. **Service layer rules** — `Service` struct in `service.go`. All business logic on the service. Constructor takes peer `*Service` types and own `*Repo`. Uses `contextx.ContextX`, not `context.Context`.
4. **Repository rules** — `repo.go` is the **only** file that imports `core/db/sqlc`. Methods named for domain operations (e.g. `GetUserAlbumRatings`), returning DTOs / domain types, never `sqlc.*` types. Repo is a concrete struct. If a service test needs a mock, the service file declares a small interface locally (Go convention: accept interfaces, return structs).
5. **Transactions** — multi-step persistence happens via `core/db.WithTx(ctx, func(tx) { ... })` at the service layer. Repo constructors take a `*sqlc.Queries` so they can be bound either to the global handle or to a transaction.
6. **Domain types** — DTOs and value objects in module-root files named by topic (e.g. `review/rating.go`, `review/state.go`). Avoid a catch-all `models.go`. Existing examples: `review/rating.go`, `review/state.go`.
7. **Allowed imports** — `core/*`; other domain modules' exported `Service` and DTO types only (via constructor injection).
8. **Forbidden imports** — other domain modules' `adapters/`; other modules' internal/non-exported packages; `sqlc` from any file other than `repo.go`.
9. **Adapters** — HTTP handlers in `adapters/http.go`. Templ files in `adapters/*.templ`, generated to `*_templ.go`. URL patterns and route registration live in `adapters/routes.go` with signature `RegisterRoutes(mux *httpx.Mux, h *HttpHandler)`. Handlers import only their own module's `*Service` (and peer `*Service` types passed via the constructor); never `repo.go`, `sqlc`, or peer modules' adapters.
10. **Module sizing & splits** — full text from spec: line count is a trigger to look, not a reason to split. Include the *Indicators worth investigating*, *Justification to split*, and *Don't split when* bullets verbatim from the spec.

Write the doc directly. Do not include placeholders or TBDs.

- [ ] **Step 4: Verify the doc answers the three required questions**

Re-read the doc you wrote and confirm an agent could answer these from it without judgment:

1. *"What files belong in this archetype?"* — Yes if the file layout tree and required/optional notes are explicit.
2. *"What imports are allowed and forbidden?"* — Yes if both lists are concrete (named packages).
3. *"Where does new code (a query, a handler, a background task) go?"*
   - Query → `repo.go`
   - Handler → `adapters/http.go`, route registered in `adapters/routes.go`
   - Background task → spec doesn't specify; if your archetype doc doesn't say either, add one sentence: background tasks live in a `task.go` at the module root and implement the `core/task.Task` interface (matches existing `feed/task.go`, `listeninghistory/task.go`).

If any of the three has a fuzzy answer, fix the doc before moving on.

- [ ] **Step 5: Commit**

```bash
git add docs/architecture/archetypes/domain-module.md
git commit -m "docs: add domain-module archetype rules"
```

---

## Task 2: Write `archetypes/external-client.md`

**Files**
- Create: `docs/architecture/archetypes/external-client.md`

- [ ] **Step 1: Read existing client modules**

Read `src/internal/musicbrainz/` (has `client.go`, `entities.go`, `service.go` — closest to the canonical shape) and `src/internal/spotify/` (has `auth.go` and `spotify.go` — diverges, will need backlog work).

- [ ] **Step 2: Write `docs/architecture/archetypes/external-client.md`**

Content must include:

1. **Purpose** — wraps a third-party API. No own domain concepts. No HTTP entrypoints (no `adapters/`).
2. **File layout** — `client.go` (SDK / HTTP client + low-level calls), `entities.go` (types from external API + conversion to internal types), `service.go` (internal-facing operations consumers use), `*_test.go`, `CLAUDE.md`. No `adapters/`. No `repo.go`. README optional — package doc comment in `client.go` is sufficient.
3. **Allowed imports** — `core/*`, vendor SDKs, stdlib.
4. **Forbidden imports** — domain modules. Clients are leaves; domain modules depend on clients, never the reverse.
5. **Why no DB / repo** — clients don't own any tables. If a 3rd-party integration needs to persist data (e.g. cached responses, OAuth tokens), persistence belongs in the consuming domain module, not the client.

- [ ] **Step 3: Verify**

Same three-questions check as Task 1. Confirm the doc answers them.

- [ ] **Step 4: Commit**

```bash
git add docs/architecture/archetypes/external-client.md
git commit -m "docs: add external-client archetype rules"
```

---

## Task 3: Write `archetypes/utility.md`

**Files**
- Create: `docs/architecture/archetypes/utility.md`

- [ ] **Step 1: Read the existing utility module**

Read `src/internal/genres/` (genres.go + data.json + tests) — currently the only utility module.

- [ ] **Step 2: Write `docs/architecture/archetypes/utility.md`**

Content must include:

1. **Purpose** — stateless: pure functions and/or embedded data. No service struct, no DB, no state.
2. **File layout** — `<name>.go`, `*_test.go`, optional embedded data (e.g. `data.json` with `//go:embed`), `CLAUDE.md`.
3. **Allowed imports** — `core/*`, stdlib, small focused libraries (e.g. fuzzy matchers).
4. **Forbidden imports** — domain modules, external clients.
5. **When to use this archetype vs. core** — utility modules live in `src/internal/<name>/` because they are *domain-shaped* utilities (e.g. genre DAG) but stateless. Code that is framework-level (HTTP, DB, context) goes in `core/` instead.

- [ ] **Step 3: Verify**

Three-questions check. Confirm.

- [ ] **Step 4: Commit**

```bash
git add docs/architecture/archetypes/utility.md
git commit -m "docs: add utility archetype rules"
```

---

## Task 4: Write `src/internal/server/CLAUDE.md` (singleton)

**Files**
- Create: `src/internal/server/CLAUDE.md`

This is the singleton documentation for the composition root. There is no archetype doc for server.

- [ ] **Step 1: Read current server.go to anchor the rules in reality**

Read `src/internal/server/server.go` end-to-end. Note: today it owns service construction (NewServices), task registration, lifecycle (Start), middleware setup, AND route registration for every module. The plan is for routes to move to per-module `adapters/routes.go` — this CLAUDE.md describes the target state, and the refactor backlog (Task 7) captures the gap.

- [ ] **Step 2: Write `src/internal/server/CLAUDE.md`**

Content:

```markdown
# server — composition root (singleton)

This directory is the **composition root** of the application. There is exactly one of it; it has no archetype because an archetype describes a category and there is only one server.

## Responsibilities

- Build all services in `NewServices(app, db)` (manual DI).
- Set up the root `*httpx.Mux` and any sub-muxes (e.g. authenticated `/app/` sub-mux with JWT middleware).
- Register cron tasks with the `core/task` task manager.
- Call each domain module's `adapters.RegisterRoutes(mux, handler)` to register routes — one call per module.
- Run lifecycle: open DB, start task manager, start HTTP listener, handle shutdown.

## Rules

- **No domain logic.** Server wires things together; it does not implement features.
- **No URL patterns** beyond mounting sub-muxes (`rootMux.Use("/app/", appMux)`). Concrete paths live in each domain module's `adapters/routes.go`.
- **Allowed imports:** every domain module, every external client, all of `core/*`. Server is the only place this is allowed.
- **No tests** in this package — the integration tested via e2e tests in `e2e/`.

## Why server is not an archetype

An archetype describes a category of modules with multiple instances and shared rules. There is exactly one server. Trying to fit it into the `utility` archetype would require carving out exceptions to utility's import rules (utility forbids importing domain modules; server requires importing all of them). A singleton is documented here directly.
```

- [ ] **Step 3: Commit**

```bash
git add src/internal/server/CLAUDE.md
git commit -m "docs: document server as composition-root singleton"
```

---

## Task 5: Write `src/internal/core/CLAUDE.md` (singleton)

**Files**
- Create: `src/internal/core/CLAUDE.md`

- [ ] **Step 1: Read the existing core README and sub-package list**

Read `src/internal/core/README.md` (the existing high-level overview) and `ls src/internal/core/` (to confirm the sub-package list: `app`, `contextx`, `cryptox`, `db`, `httpx`, `sqlx`, `task`, `templates`, `timex`, `utils`).

- [ ] **Step 2: Write `src/internal/core/CLAUDE.md`**

Content:

```markdown
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
```

- [ ] **Step 3: Commit**

```bash
git add src/internal/core/CLAUDE.md
git commit -m "docs: document core as shared-infrastructure singleton"
```

---

## Task 6: Write `docs/architecture/README.md`

**Files**
- Create: `docs/architecture/README.md`

This is the front door of the architecture docs. It indexes the archetypes, points at the singletons, and explains the layered-CLAUDE.md mechanism.

- [ ] **Step 1: Confirm previous tasks landed**

Run: `ls docs/architecture/archetypes/ src/internal/server/CLAUDE.md src/internal/core/CLAUDE.md`
Expected: `domain-module.md  external-client.md  utility.md` plus the two singleton CLAUDE.md files exist.

- [ ] **Step 2: Write `docs/architecture/README.md`**

Content:

```markdown
# Wax Architecture

This directory documents the architectural rules for `src/internal/`. Most directories under `src/internal/` are classified into one of three archetypes; two directories are singletons documented in their own `CLAUDE.md`.

## Archetypes

| Archetype | Doc | Examples |
|---|---|---|
| Domain module | [archetypes/domain-module.md](archetypes/domain-module.md) | `library`, `review`, `tags`, `notes`, `auth`, ... |
| External client | [archetypes/external-client.md](archetypes/external-client.md) | `spotify`, `musicbrainz`, `discogs` |
| Utility | [archetypes/utility.md](archetypes/utility.md) | `genres` |

## Singletons

Two directories are exactly-one-of-them. Their rules live next to the code:

- **`server/`** — composition root. Builds services, sets up middleware and sub-muxes, calls each domain module's `RegisterRoutes`, runs lifecycle. See [`src/internal/server/CLAUDE.md`](../../src/internal/server/CLAUDE.md).
- **`core/`** — shared infrastructure. Framework-level sub-packages used by 2+ modules. See [`src/internal/core/CLAUDE.md`](../../src/internal/core/CLAUDE.md).

A singleton is *not* an archetype: archetypes describe categories with multiple instances. Trying to fit `server` into `utility` (or any other archetype) would require carving out exceptions to that archetype's import rules.

## Encoding mechanism

Architectural rules are encoded as a layered set of `CLAUDE.md` files that Claude Code auto-loads when working in a relevant subtree:

- **Root `.claude/CLAUDE.md`** — points at this directory.
- **Per-directory `src/internal/<dir>/CLAUDE.md`** — declares the directory's archetype (or, for singletons, documents rules directly) plus any module-specific notes.
- **Archetype docs in `archetypes/`** — full rules for each category.

## Listing archetypes at a glance

To see which archetype every directory is classified as:

```bash
grep -h "^# " src/internal/*/CLAUDE.md
```

(There is no separate module registry — that would duplicate what each `CLAUDE.md` already declares and would drift the same way the wiki did.)

## Refactor backlog

[`refactor-backlog.md`](refactor-backlog.md) lists, per directory, the compliance gaps and structural notes needed to bring the code into line with these rules. The refactor itself is downstream work, not part of this documentation effort.
```

- [ ] **Step 3: Commit**

```bash
git add docs/architecture/README.md
git commit -m "docs: add architecture README indexing archetypes and singletons"
```

---

## Task 7: Write `docs/architecture/refactor-backlog.md`

**Files**
- Create: `docs/architecture/refactor-backlog.md`

A heuristic pass over each directory under `src/internal/`. Each entry has two parts: compliance gaps and structural notes. Entries are 1–3 sentences each.

- [ ] **Step 1: Inventory each module**

For each directory in `src/internal/`, run a quick read:

```bash
ls src/internal/<dir>/
wc -l src/internal/<dir>/*.go
head -40 src/internal/<dir>/service.go      # (when present)
```

For each, note:
- **Files present vs. archetype layout** (missing `repo.go`, missing `adapters/`, missing `routes.go`, missing `README.md`, missing `service.go`, etc.)
- **Imports of `sqlc` outside `repo.go`** — `grep -l "core/db/sqlc" src/internal/<dir>/*.go`
- **Routes still declared in `server/server.go`** — they all are today; flag this in every domain module entry.
- **Responsibility coherence** — read `service.go`, ask: can I describe this in one sentence without "and"?

- [ ] **Step 2: Write `docs/architecture/refactor-backlog.md`**

Use this structure:

```markdown
# Refactor Backlog

Per-directory work needed to bring `src/internal/` into compliance with the archetype rules. Entries have two parts:

1. **Compliance** — missing files, files in the wrong place, forbidden imports.
2. **Structural** — splits, merges, moves, or renames. Optional; only present where there's a real issue.

Entries are not prescriptive about *how* to do the work — that's the refactor's job. They name the gap and the rationale.

---

## auth (domain)

**Compliance:** No `service.go`, no `repo.go`, no `adapters/` — everything lives in `http.go` and `components.templ` at the module root. Should grow an `auth.Service` (JWT issuance, login orchestration, OAuth callback handling), with HTTP moved to `adapters/http.go` + `adapters/routes.go` + `adapters/components.templ`.

**Structural:** None — `auth` is a coherent responsibility once it's properly shaped.

---

## core (singleton)

No archetype gaps. Already in shape. Structural review: confirm every sub-package is used by 2+ modules; flag any single-consumer sub-package as a candidate for moving back to its consumer.

---

## discogs (external-client)

**Compliance:** Has `client.go`, `entities.go`, `service.go`, `genres.go`, `genres_test.go`. Missing `CLAUDE.md` (added by Task 9). The `genres.go` here looks like domain logic (fuzzy genre matching against the discogs response) — verify it belongs in `discogs` rather than the `genres` utility module.

**Structural:** Possibly merge `discogs/genres.go` into the `genres` utility if it's general genre-matching logic; or keep here if it's discogs-specific normalization.

---

## feed (domain)

**Compliance:** Has `service.go` and `task.go`. Missing `repo.go` (sqlc imported directly in `service.go`). Has no HTTP, so no `adapters/` needed — confirmed correct. Missing `README.md`.

**Structural:** None.

---

## genres (utility)

**Compliance:** Has `genres.go`, `genres_test.go`, `data.json`. Missing `CLAUDE.md` (added by Task 11). Otherwise compliant with the utility archetype.

**Structural:** None — coherent responsibility (the genre DAG).

---

## labels (domain)

**Compliance:** Empty `adapters/` directory. No `service.go`, no `repo.go`. Module is essentially a stub; either flesh it out as a domain module or remove until needed.

**Structural:** Decide intent — if labels is on the roadmap, leave the stub and add a backlog item; if not, remove the directory.

---

## library (domain)

**Compliance:** Extract `repo.go`; the `sqlc` import disappears from `service.go`. Move route registration from `server/server.go` into `library/adapters/routes.go`.

**Structural:** `service.go` is 1009 lines and currently owns albums, artists, tracks, releases, formats, and user-collection state. Artists and tracks are independently meaningful concepts that other modules could plausibly reference directly. Propose extracting `artists` and `tracks` as separate domain modules; `library` keeps the user-collection aggregate (the user's relationship to albums, releases, formats).

---

## listeninghistory (domain)

**Compliance:** Has `service.go` and `task.go`. Missing `repo.go`. No HTTP, so no `adapters/` — correct.

**Structural:** None.

---

## musicbrainz (external-client)

**Compliance:** Has `client.go`, `entities.go`, `service.go`. Closest to the canonical client shape. Missing `CLAUDE.md` (added by Task 10).

**Structural:** None.

---

## notes (domain)

**Compliance:** Has `service.go`, `service_test.go`, `adapters/`. Missing `repo.go`. Move route registration from `server/server.go` into `notes/adapters/routes.go`.

**Structural:** None.

---

## review (domain)

**Compliance:** Has `service.go`, `rating.go`, `state.go`, `README.md`, tests, `adapters/`. Closest to the documented shape. Missing `repo.go`. Move route registration into `review/adapters/routes.go`.

**Structural:** None.

---

## server (singleton)

**Compliance:** Currently owns route registration for every module — about 80% of `server.go` is route-table boilerplate. Move per-module route declarations into each domain module's `adapters/routes.go`. Server retains: `NewServices`, mux + middleware setup, task registration, lifecycle, and a list of `RegisterRoutes` calls.

**Structural:** Split `server.go` into focused files once routes move out: `services.go` (DI), `start.go` (lifecycle + middleware + RegisterRoutes calls). No need to over-split.

---

## spotify (external-client)

**Compliance:** Has `auth.go` and `spotify.go` — diverges from the canonical client layout (no `client.go`/`entities.go`/`service.go` split). Reorganize: `client.go` for the low-level Spotify client, `entities.go` for type wrappers, `service.go` for `Service` + `AuthService`. Or keep `auth.go` separate if Auth genuinely deserves its own file.

**Structural:** None — responsibility (Spotify integration) is coherent.

---

## tags (domain)

**Compliance:** Has `service.go` and `adapters/`. Missing `repo.go`. Move route registration into `tags/adapters/routes.go`. Missing `README.md`.

**Structural:** None.

---

## user (domain)

**Compliance:** Only has `service.go`. Missing `repo.go`. No `adapters/` — likely correct (user has no HTTP entrypoints today; auth handles the user-facing surface).

**Structural:** None — but verify whether OAuth tokens should live here (currently?) or in `auth` with the new shape.
```

Adjust each entry based on what you actually find when inventorying. The above is the seed; correct it where reality differs.

- [ ] **Step 3: Commit**

```bash
git add docs/architecture/refactor-backlog.md
git commit -m "docs: add architecture refactor backlog"
```

---

## Task 8: Write per-module `CLAUDE.md` for the 9 domain modules

**Files**
- Create: `src/internal/auth/CLAUDE.md`
- Create: `src/internal/feed/CLAUDE.md`
- Create: `src/internal/labels/CLAUDE.md`
- Create: `src/internal/library/CLAUDE.md`
- Create: `src/internal/listeninghistory/CLAUDE.md`
- Create: `src/internal/notes/CLAUDE.md`
- Create: `src/internal/review/CLAUDE.md`
- Create: `src/internal/tags/CLAUDE.md`
- Create: `src/internal/user/CLAUDE.md`

Each file follows the same template; contents differ only in the module name and module-specific notes.

- [ ] **Step 1: Confirm the archetype doc landed**

Run: `test -f docs/architecture/archetypes/domain-module.md && echo OK`
Expected: `OK`.

- [ ] **Step 2: For each domain module, identify 1–3 module-specific notes worth flagging**

For each of the 9 modules, briefly read its current files (`service.go` and any peer files). Look for things that would not be obvious from the archetype rules alone:

- Cross-module dependencies (e.g. *"depends on `spotify.Service` for syncing the saved-albums feed"*)
- Background tasks (e.g. *"registers `SyncSpotifyFeedTask` cron task"*)
- Currently-non-compliant state (e.g. *"NOT YET COMPLIANT: no `repo.go`; sqlc imported directly in `service.go`. See refactor backlog."*)
- Public API quirks (e.g. *"`auth.Service` is invoked by the `/login` and `/spotify/callback` routes only"*)

It's fine for a module to have only one note. Don't pad.

- [ ] **Step 3: Write each `CLAUDE.md` using this template**

```markdown
# <module> — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- <note 1>
- <note 2>  (optional)
- <note 3>  (optional)
```

Notes:
- The relative path `../../../docs/architecture/archetypes/domain-module.md` is correct from `src/internal/<module>/`.
- For modules without `adapters/` (`feed`, `listeninghistory`, `user`), add a note: *"No HTTP entrypoints currently — `adapters/` is intentionally absent."*
- For modules currently non-compliant, add a note: *"NOT YET COMPLIANT: see `docs/architecture/refactor-backlog.md` for gaps."*

Example for `review`:

```markdown
# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Closest existing module to the canonical domain-module shape.
- Pure-logic types live in `rating.go` and `state.go` (good example of "split by topic").
- NOT YET COMPLIANT: missing `repo.go`; routes still declared in `server/server.go`. See `docs/architecture/refactor-backlog.md`.
```

Example for `feed`:

```markdown
# feed — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- No HTTP entrypoints — `adapters/` is intentionally absent.
- Owns the cron task `SyncStaleSpotifyFeedsTask` (see `task.go`); registered by `server/`.
- NOT YET COMPLIANT: missing `repo.go` and `README.md`. See `docs/architecture/refactor-backlog.md`.
```

- [ ] **Step 4: Verify all 9 files exist**

Run:
```bash
for d in auth feed labels library listeninghistory notes review tags user; do
  test -f src/internal/$d/CLAUDE.md && echo "OK $d" || echo "MISSING $d"
done
```
Expected: nine `OK` lines.

- [ ] **Step 5: Commit**

```bash
git add src/internal/auth/CLAUDE.md src/internal/feed/CLAUDE.md src/internal/labels/CLAUDE.md src/internal/library/CLAUDE.md src/internal/listeninghistory/CLAUDE.md src/internal/notes/CLAUDE.md src/internal/review/CLAUDE.md src/internal/tags/CLAUDE.md src/internal/user/CLAUDE.md
git commit -m "docs: add per-module CLAUDE.md for domain modules"
```

---

## Task 9: Write per-module `CLAUDE.md` for the 3 external-client modules

**Files**
- Create: `src/internal/discogs/CLAUDE.md`
- Create: `src/internal/musicbrainz/CLAUDE.md`
- Create: `src/internal/spotify/CLAUDE.md`

- [ ] **Step 1: Read each external client briefly**

Read `service.go` (or equivalent) for each. Note module-specific quirks (e.g. spotify has both `Service` and `AuthService`).

- [ ] **Step 2: Write each `CLAUDE.md`**

Template:

```markdown
# <module> — external client

Rules: ../../../docs/architecture/archetypes/external-client.md

Module-specific notes:
- <note 1>
- <note 2>  (optional)
```

Specific content guidance:

- **`discogs`** — flag that `genres.go` may belong in the `genres` utility module (see refactor backlog).
- **`musicbrainz`** — flag this as the closest-to-canonical example for the archetype.
- **`spotify`** — flag the divergent layout (`auth.go` + `spotify.go` instead of `client.go` + `entities.go` + `service.go`) and that it exposes both `Service` and `AuthService`.

- [ ] **Step 3: Commit**

```bash
git add src/internal/discogs/CLAUDE.md src/internal/musicbrainz/CLAUDE.md src/internal/spotify/CLAUDE.md
git commit -m "docs: add per-module CLAUDE.md for external clients"
```

---

## Task 10: Write `src/internal/genres/CLAUDE.md`

**Files**
- Create: `src/internal/genres/CLAUDE.md`

- [ ] **Step 1: Write the file**

```markdown
# genres — utility

Rules: ../../../docs/architecture/archetypes/utility.md

Module-specific notes:
- Currently the only utility module in the codebase.
- Builds a genre DAG from `data.json` (Wikidata-derived) using `//go:embed`. Provides fuzzy lookup via `lithammer/fuzzysearch`.
- Stateless; no constructor — callers use package-level functions (or `Load()` returning the DAG).
```

- [ ] **Step 2: Commit**

```bash
git add src/internal/genres/CLAUDE.md
git commit -m "docs: add CLAUDE.md for genres utility module"
```

---

## Task 11: Update root `.claude/CLAUDE.md`

**Files**
- Modify: `.claude/CLAUDE.md`

Replace the existing "Architecture Patterns" section with a short pointer. Preserve everything else (Code Generation, Development, Testing, Documentation sections).

- [ ] **Step 1: Read the current file**

Read `.claude/CLAUDE.md`. Confirm the "Architecture Patterns" section is the one starting with `## Architecture Patterns` and ending before `## Development`.

- [ ] **Step 2: Replace the "Architecture Patterns" section**

Replace from `## Architecture Patterns` through (exclusive of) `## Development` with:

```markdown
## Architecture Patterns

`src/internal/` is organized by archetype. Every directory under `src/internal/` has a `CLAUDE.md` declaring its archetype (or, for `server` and `core`, documenting singleton rules). When working in a module, the relevant rules will auto-load with that directory's `CLAUDE.md`.

Full rules: [`docs/architecture/`](../docs/architecture/).

For agents adding new code:
- New code under `src/internal/<module>/` must follow the rules of that module's archetype. Read the module's `CLAUDE.md` first.
- New modules: pick an archetype before writing code. If unsure, see [`docs/architecture/README.md`](../docs/architecture/README.md).
```

- [ ] **Step 3: Verify the rest of the file is intact**

Run: `grep -c "^## " .claude/CLAUDE.md`
Expected: 5 (Code Generation, Architecture Patterns, Development, Testing, Documentation).

Also confirm the Code Generation and Development sections still mention `task build/templ`, `task build/sqlc`, etc.

- [ ] **Step 4: Commit**

```bash
git add .claude/CLAUDE.md
git commit -m "docs: replace inline architecture rules with pointer to docs/architecture/"
```

---

## Task 12: Trim `docs/wiki/pages/architecture.md`

**Files**
- Modify: `docs/wiki/pages/architecture.md`

Keep the stack table and the high-level module list. Replace the "Key Patterns" section with a one-line pointer.

- [ ] **Step 1: Read the current file**

Read `docs/wiki/pages/architecture.md`. Confirm the "Key Patterns" section starts with `## Key Patterns` and ends before `## Deployment`.

- [ ] **Step 2: Replace the "Key Patterns" section**

Replace from `## Key Patterns` through (exclusive of) `## Deployment` with:

```markdown
## Key Patterns

Architectural rules — module archetypes, allowed/forbidden imports, file layout, repo and adapter conventions — live in [`docs/architecture/`](../../architecture/). The wiki summarizes the system; the architecture docs are the authoritative source for rules.

```

(Note: the relative link from `docs/wiki/pages/architecture.md` to `docs/architecture/` is `../../architecture/`.)

- [ ] **Step 3: Verify**

Run:
```bash
test -f docs/wiki/pages/architecture.md && grep -q "docs/architecture\|/architecture/" docs/wiki/pages/architecture.md && echo OK
```
Expected: `OK`.

Also confirm the stack table at the top of the file is still intact.

- [ ] **Step 4: Commit**

```bash
git add docs/wiki/pages/architecture.md
git commit -m "docs: replace wiki architecture key-patterns with pointer to docs/architecture/"
```

---

## Task 13: Sanity check — fresh-agent dry run

**Files**
- None (this is a verification task)

This is the integration test for the project. Confirm that an agent given only `.claude/CLAUDE.md` would land code in the right place.

- [ ] **Step 1: List all 15 per-directory CLAUDE.md files**

Run:
```bash
ls src/internal/*/CLAUDE.md | wc -l
```
Expected: `15`.

Run:
```bash
grep -h "^# " src/internal/*/CLAUDE.md | sort
```
Expected: 15 lines, each like `# <name> — <archetype>` or `# <name> — <singleton designation>`. No duplicates. No missing modules.

- [ ] **Step 2: Confirm archetype docs are reachable**

Run:
```bash
ls docs/architecture/archetypes/
```
Expected: `domain-module.md  external-client.md  utility.md`.

Run:
```bash
ls docs/architecture/
```
Expected: `README.md  archetypes  refactor-backlog.md`.

- [ ] **Step 3: Run the three required questions test against `archetypes/domain-module.md`**

Re-read `docs/architecture/archetypes/domain-module.md`. Confirm an agent could answer, without judgment:

1. *"I'm adding a new SQL query for the library module. What file?"* — should land at `src/internal/library/repo.go` (a method on `Repo`, returning a domain type).
2. *"I'm adding a new GET endpoint to the review module. What files?"* — handler method in `review/adapters/http.go`; route line in `review/adapters/routes.go`; templ if HTML response.
3. *"I'm adding a new background task for syncing something."* — `task.go` at the module root, implementing `core/task.Task`, registered in `server/server.go` via `taskManager.RegisterCronTask`.

If any of these is ambiguous in the doc, fix the doc.

- [ ] **Step 4: Verify root `.claude/CLAUDE.md` no longer contains stale architecture rules**

Run:
```bash
grep -i "service.go\|adapters/\|repo.go" .claude/CLAUDE.md
```
Expected: no matches (or only inside the pointer blurb). Architecture detail should now live in `docs/architecture/`.

- [ ] **Step 5: Final commit (if anything was tweaked in step 3)**

If nothing changed, skip this step. If you fixed a doc:

```bash
git add docs/architecture/archetypes/domain-module.md
git commit -m "docs: tighten domain-module archetype rules to pass sanity check"
```

- [ ] **Step 6: Verify the full set of expected files**

Run:
```bash
ls docs/architecture/README.md \
   docs/architecture/refactor-backlog.md \
   docs/architecture/archetypes/domain-module.md \
   docs/architecture/archetypes/external-client.md \
   docs/architecture/archetypes/utility.md \
   src/internal/*/CLAUDE.md \
   .claude/CLAUDE.md
```
Expected: every file listed (no errors).

- [ ] **Step 7: Verify the project leaves no stranded `.go` changes**

Run: `git status --short`
Expected: clean (or only the pre-existing `static/public/main.css` modification, if it was there at the start).

The project is complete when every checkbox above is checked.

---

## Self-Review Notes (for the plan author)

After writing this plan, I checked it against the spec:

- **Spec coverage:**
  - Spec section *Goals & Non-Goals* — Tasks 1–13 deliver the goals; non-goals (no `.go` changes, no hooks, no wiki restructure beyond the pointer) are baked into task scopes and the final verification.
  - Spec section *Background* — context, used implicitly by Task 7 (refactor backlog). Not a deliverable.
  - Spec section *Module Archetypes* (3 archetypes + sizing rules) — Tasks 1, 2, 3.
  - Spec section *Singletons* — Tasks 4 (server), 5 (core), 6 (README cross-link).
  - Spec section *Module Classification* — Tasks 8, 9, 10 (per-module CLAUDE.md files).
  - Spec section *Encoding (file layout)* — Tasks 1–6, 8–11 produce the new files; Task 11 updates root `.claude/CLAUDE.md`; Task 12 updates wiki.
  - Spec section *Why no module registry* — Task 6 covers via the grep one-liner in the README.
  - Spec section *Refactor Backlog* — Task 7.
  - Spec section *Definition of Done* — Task 13 verifies all 7 acceptance criteria.

- **Placeholder scan:** The plan asks the implementer to fill in module-specific notes after reading the actual code (Task 8 step 2, Task 9 step 1, Task 7 step 1). These are not placeholders in the doc deliverable — they're a fact-finding step the implementer must do. The deliverables themselves contain no TBDs.

- **Type consistency:** `RegisterRoutes(mux *httpx.Mux, h *HttpHandler)` is the consistent signature throughout. `Repo` is consistently a concrete struct (interfaces declared locally in service files when needed). `core/db.WithTx(ctx, func(tx) { ... })` is consistent. Path conventions (`src/internal/<dir>/`) match across tasks.
