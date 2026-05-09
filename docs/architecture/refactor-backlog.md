# Refactor Backlog

Per-directory work needed to bring `src/internal/` into compliance with the archetype rules at [`docs/architecture/`](.). Entries have two parts:

1. **Compliance** — missing files, files in the wrong place, forbidden imports.
2. **Structural** — splits, merges, moves, or renames. Optional; only present where there's a real issue.

Entries are not prescriptive about *how* to do the work — that's the refactor's job. They name the gap and the rationale.

---

## auth (domain)

**Compliance:** No `service.go`, no `repo.go`, no `adapters/` — everything lives in `http.go` and `components.templ` at the module root, and `http.go` imports `core/db/models` directly. Should grow an `auth.Service` (JWT issuance, login orchestration, OAuth callback handling) plus `repo.go`, with HTTP moved to `adapters/http.go` + `adapters/routes.go` + `adapters/components.templ`. Routes for `/`, `/logout`, and `/spotify/callback` are registered in `server/server.go` and should move to `auth/adapters/routes.go`.

**Structural:** None — `auth` is a coherent responsibility once it's properly shaped.

---

## core (singleton)

No archetype gaps. Already in shape per `src/internal/core/CLAUDE.md`. Structural review: confirm every sub-package is used by 2+ modules; flag any single-consumer sub-package as a candidate for moving back to its consumer.

---

## discogs (external-client)

**Compliance:** Has `client.go`, `entities.go`, `service.go`, `genres.go`, `genres_test.go`. Missing `CLAUDE.md` (added by Task 9). The `genres.go` here wraps the `genres` utility with Discogs-specific compound-term splitting (e.g. "Funk / Soul", "Folk, World, & Country") and item resolution against `SearchItem` — that's Discogs-shaped, not general genre logic.

**Structural:** Keep `genres.go` here — it's a thin Discogs-specific adapter over the `genres` utility, not general fuzzy matching that other modules would reuse.

---

## feed (domain)

**Compliance:** Has `service.go` and `task.go`. Missing `repo.go` (sqlc imported directly in `service.go`). Has no HTTP, so no `adapters/` needed — confirmed correct. Missing `README.md`.

**Structural:** None.

---

## genres (utility)

**Compliance:** Has `genres.go`, `genres_test.go`, `data.json`. Missing `CLAUDE.md` (added by Task 10). Otherwise compliant with the utility archetype.

**Structural:** None — coherent responsibility (the genre DAG).

---

## labels (domain)

**Compliance:** Empty `adapters/` directory and nothing else — no `service.go`, no `repo.go`, no files at all. The module is not wired into `server.go`. Either flesh it out as a domain module or remove until needed.

**Structural:** Decide intent — if labels is on the roadmap, leave the stub and add a backlog item; if not, remove the directory.

---

## library (domain)

**Compliance:** Done — `repo.go` extracted, routes in `library/adapters/routes.go`, business logic lifted out of `adapters/formats.go`, `README.md` added. Repo currently wraps rating-table queries (`GetLatestUserAlbumRating`, `GetUserAlbumRatingLog`, etc.) that belong in `review`; flagged with `// TODO` in `repo.go` for follow-up.

**Structural:** Topic-file split done — `service.go` (Service + methods only), `album.go`, `release.go`, `view.go`. Module-level splits (extracting `artists`/`tracks`) remain unjustified.

Sleeve notes UI moves entirely into `library/adapters/` (display + editor + handlers). Today's `notes/adapters/notes.templ` and `notes/adapters/http.go` both import `library/adapters` — a forbidden cross-adapter import. The album view is library's territory; sleeve notes are an inline part of it, so library renders them and calls `notes.Service` for persistence/markdown. Move `SleeveNotesEditor` into `library/adapters/sleeve_notes.templ`, move the three sleeve-note handlers into `library/adapters/http.go`, and re-prefix the routes (e.g. `/app/library/albums/{id}/sleeve-notes/{editor,view}` and PUT). See the `notes` entry for the corresponding cleanup.

---

## listeninghistory (domain)

**Compliance:** Has `service.go` and `task.go`. Missing `repo.go` (sqlc imported directly in `service.go`). No HTTP, so no `adapters/` — correct.

**Structural:** None.

---

## musicbrainz (external-client)

**Compliance:** Has `client.go`, `entities.go`, `service.go`. Closest to the canonical client shape. Missing `CLAUDE.md` (added by Task 9).

**Structural:** None.

---

## notes (domain)

**Compliance:** Has `service.go`, `service_test.go`, `adapters/http.go`, `adapters/notes.templ`. Missing `repo.go` (sqlc imported directly in `service.go`). The current `adapters/` cross-imports `library/adapters` — a forbidden adapter-to-adapter import — and exists solely to render sleeve notes, which are part of library's album view.

**Structural:** Drop `notes/adapters/` entirely. The sleeve-notes UI (display, editor, handlers) moves to `library/adapters/`; `notes` becomes a pure data + markdown-rendering service like `feed`/`listeninghistory`/`user` — `service.go`, `repo.go`, tests, no `adapters/`. Library's handlers depend on `notes.Service` for upsert/fetch and `notes.RenderMarkdown` for HTML rendering. Routes for sleeve notes get re-prefixed under `/app/library/...` since library owns the view. Naming aside: the persistence type is `AlbumNote` and the UI label is "sleeve note" — these refer to one concept; keep the persistence name and treat "sleeve note" as a UI string only.

---

## review (domain)

**Compliance:** Has `service.go`, `rating.go`, `state.go`, `rating_test.go`, `state_test.go`, `README.md`, and `adapters/`. Closest to the documented shape. Missing `repo.go` (sqlc imported directly in `service.go`). Move route registration (`/app/review/...`) into `review/adapters/routes.go`.

**Structural:** None.

---

## server (singleton)

**Compliance:** Currently owns route registration for every module — the bulk of `server.go` is route-table boilerplate (`appMux.Handle(...)` for library, tags, notes, review, plus `rootMux.Handle(...)` for auth). Move per-module route declarations into each domain module's `adapters/routes.go`. Server retains: `NewServices`, mux + middleware setup, task registration, lifecycle, and a list of `RegisterRoutes` calls.

**Structural:** Split `server.go` into focused files once routes move out: `services.go` (DI), `start.go` (lifecycle + middleware + RegisterRoutes calls). No need to over-split.

---

## spotify (external-client)

**Compliance:** Has `auth.go` and `spotify.go` — diverges from the canonical client layout (no `client.go`/`entities.go`/`service.go` split). Reorganize into `client.go` for the low-level Spotify client, `entities.go` for type wrappers (e.g. the `SavedAlbum` alias), `service.go` for `Service`, and keep `auth.go` for `AuthService` since it's a meaningfully separate concern (OAuth flow, callback handling). Missing `CLAUDE.md` (added by Task 9).

**Structural:** None — responsibility (Spotify integration) is coherent.

---

## tags (domain)

**Compliance:** Has `service.go` and `adapters/`. Missing `repo.go` (sqlc imported directly in `service.go`). Move route registration (`/app/tags/...`) into `tags/adapters/routes.go`. Missing `README.md`.

**Structural:** None.

---

## user (domain)

**Compliance:** Only has `service.go`. Missing `repo.go` (sqlc imported directly). No `adapters/` — likely correct (user has no HTTP entrypoints today; auth handles the user-facing surface).

**Structural:** None — but verify whether the encrypted Spotify refresh token (currently a `UserDTO` field plus `cryptox` decrypt method) belongs here or in `auth` once the auth module gets its proper shape.
