# Domain Module

## Purpose

A domain module owns a slice of the application's domain end-to-end: business logic, persistence, and (optionally) HTTP entrypoints. Most modules under `src/internal/` are domain modules. Each one is the single home for the rules and data that belong to its slice — peer modules consume it only through its exported `Service` and DTO types.

## File layout

```
src/internal/<module>/
├── service.go          # Service struct + business logic — required
├── repo.go             # ONLY file allowed to import core/db/sqlc — required if the module owns persistence
├── <package>.go        # domain types, view models, pure helpers — optional, default name matches the package (e.g. library/library.go)
├── task.go             # background tasks (core/task.Task) — optional
├── *_test.go           # tests live next to the file under test
├── README.md           # required
├── CLAUDE.md           # required; declares archetype
└── adapters/           # required only if the module has HTTP entrypoints
    ├── http.go         # HttpHandler struct + handler methods
    ├── routes.go       # RegisterRoutes(mux *httpx.Mux, h *HttpHandler)
    ├── *.templ         # templ source
    └── *_templ.go      # generated; do not edit by hand
```

Required: `service.go`, `README.md`, `CLAUDE.md`. `repo.go` is required if the module owns persistence (the common case); a module that delegates all persistence to peer services can omit it. Everything else is optional. `adapters/` exists if and only if the module exposes HTTP routes.

## Service layer

- The `Service` struct lives in `service.go`. All business logic is a method on `*Service`.
- The constructor takes peer domain modules' `*Service` types and the module's own `*Repo`. Wiring happens in `server/`.
- Methods take `contextx.ContextX`, never `context.Context`. Extract the user ID with `ctx.UserId()`.
- Services hold concrete `*Repo`. If a service test needs to mock the repo, the service file declares a small interface locally next to the consumer (Go convention: accept interfaces, return structs).

## Repository layer

- `repo.go` is the **only** file in the module allowed to import `core/db/sqlc`.
- Methods are named for domain operations (e.g. `GetUserAlbumRatings`, `InsertReview`) and return DTOs / domain types — never `sqlc.*` types. SQLC types do not appear in any signature outside `repo.go`.
- `Repo` is a concrete struct. Its constructor takes a `*sqlc.Queries` so it can be bound either to the global handle or to a transaction (see *Transactions* below).

## Transactions

- Multi-step persistence happens at the service layer via `core/db.WithTx`:

  ```go
  err := s.db.WithTx(func(tx *db.DB) error {
      // build a transactional repo bound to tx.Queries() and call it
  })
  ```

- Repo constructors take `*sqlc.Queries` (not `*db.DB`) so the same repo type works against the global handle or a transaction.

- A repo method does **one** persistence operation. If correctness depends on multiple writes landing atomically — including cross-domain side effects on other modules' tables — that orchestration belongs at the service layer, not bundled in a repo method. The service composes single-purpose repo calls inside `WithTx`; each repo method stays callable on its own.

- **Exception:** if a repo method genuinely must run inside a transaction, encode that in its signature with a type that callers can only obtain from a tx context (e.g. a `*db.Tx` argument). Invariants on repo methods are enforced by the type system, not by comments telling callers what to do.

## Domain types

- **Default: one topic file named after the package** (e.g. `library/library.go`, `tags/tags.go`). It holds all DTOs, value objects, view models, and pure helpers — methods on those types belong in the same file as the type.
- Do not create a `models.go` file.
- Types that cross module boundaries (e.g. consumed by another module's `Service`) must be exported.

### When to use multiple topic files

Default to one. Split into multiple topic files **only when all** of these hold:

- The two areas are **distinct concepts**, not two views of the same aggregate.
- They **share no types** — neither references the other's types in its own definitions.
- **No methods cross them** — there is no method that needs to reach across the split.

If any of those fail, it's one topic, one file.

Size is a *signal* that splitting might be worth investigating, not a *reason* to split. A 500-line topic file that meets the rules above stays one file.

Canonical example of a justified split: `review/rating.go` (rating values, scoring questions, labels) and `review/state.go` (rating-state machine — snoozing, rerate timing). They are genuinely independent concepts: a rating is a value, a state is a workflow; they share no types and no methods cross them.

Canonical example of a *wrong* split: separating `library` into `album.go` + `release.go` + `view.go`. Albums, releases, and the dashboard slice operations are all parts of one aggregate cluster — they share types (`AlbumDTO` carries `[]ReleaseDTO`; `AlbumDTOs` is a slice of `AlbumDTO`) and methods cross them (`AlbumDTOs.SortByDate` calls into `ReleaseDTOs.OldestAddedAtDate`). One `library.go` file is correct.

## Allowed imports

- `core/*` sub-packages.
- Other domain modules' **exported** `Service` types and DTO / value-object types, injected via the constructor.
- Stdlib and small focused third-party libraries (e.g. `github.com/google/uuid`).

## Forbidden imports

- `core/db/sqlc` from any file other than `repo.go`.
- Other domain modules' `adapters/` packages (e.g. `library/adapters` may not be imported by `review`).
- Other domain modules' unexported types or non-`Service` internal helpers.
- External-client modules from anywhere — and even where allowed (in domain modules that explicitly own those integrations), only via constructor injection of the client's `*Service`.

## Adapters

- HTTP handlers are methods on `HttpHandler` in `adapters/http.go`. The struct's fields are the peer `*Service` types it needs; a constructor `NewHttpHandler(...)` takes them.
- Templ components live in `adapters/*.templ` and are generated to `*_templ.go` by `task build/templ`. Never edit `*_templ.go` by hand.
- URL patterns and route registration live in `adapters/routes.go`. Signature:

  ```go
  func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
      mux.Handle("GET /reviews/{id}", httpx.HandlerFunc(h.GetReview))
      // ...
  }
  ```

  `server/` passes the appropriate mux (root mux for public routes, app sub-mux for authenticated routes). The module decides which paths and methods bind to which handler methods.
- Adapters import their own module's `*Service` and peer modules' `*Service` / DTO types **only**. They do not import `repo.go`, `sqlc`, or peer modules' `adapters/` packages.
- Error responses go through `httpx.HandleErrorResponse`. Responses are HTML fragments for HTMX consumption.

## Background tasks

- Background tasks live in `task.go` at the module root.
- They implement the `core/task.Task` interface (`Run`, `Schedule`, `Name`).
- Constructors are named `NewXxxTask(service *Service, ...) task.Task`.
- Tasks call into the module's own `*Service`, never directly into the repo.
- For working examples, look at existing `task.go` files in `src/internal/`.

## Where new code goes

| Change | File |
|---|---|
| New SQL query | `repo.go` (and add the `.sql` file under `db/queries/`, then `task build/sqlc`) |
| New business-logic method | `service.go` |
| New domain type / pure function | `<package>.go` at module root (or the existing topic file if there's a justified split) |
| New HTTP handler | `adapters/http.go` |
| New URL route | `adapters/routes.go` |
| New templ component | `adapters/<name>.templ`, then `task build/templ` |
| New background task | `task.go` (implements `core/task.Task`) |

## Module sizing & splits

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
