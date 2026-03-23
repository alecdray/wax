# Sleeve Notes — Research

## Summary

"Sleeve notes" refers to free-text notes a user can write about an album — personal impressions, context, or anything they want to remember — analogous to the liner notes printed inside an album sleeve. The feature is a distinct concept from the numeric rating and its attached `note` field, which is a short annotation on a specific rating log entry. Sleeve notes would be a persistent, standalone free-text field per user/album, editable at any time and independent of the rating workflow. This research maps the current codebase to determine where sleeve notes fit, what new infrastructure is needed, and what patterns to follow.

## Relevant Code

| Path | Role |
|------|------|
| `db/schema.sql` | Canonical DB schema. Currently has `album_rating_log` (with `note text`) but no standalone notes table. |
| `db/migrations/20260313204129_album_rating_log.sql` | Most recent migration — replaced `album_ratings` with `album_rating_log`; the old `album_ratings.review` column was migrated here as `note`. |
| `db/queries/album_ratings.sql` | All SQL for rating log: insert, delete, get latest, get log. No notes-specific query exists. |
| `src/internal/review/service.go` | `Service` with `AddRating`, `DeleteRatingEntry`, `GetRatingLog`. `AlbumRatingDTO` holds `Rating *float64` and `Note *string` (the per-entry note). |
| `src/internal/review/rating.go` | Rating recommender logic, labels, key. No notes logic. |
| `src/internal/review/adapters/http.go` | HTTP handlers for rating flow. References `libraryAdapters` for OOB swaps. |
| `src/internal/review/adapters/rating.templ` | Rating modal template — includes a `note` textarea in the rating confirm form. |
| `src/internal/library/service.go` | `AlbumDTO` (holds `Rating`, `RatingLog`, `Tags`, `LastPlayedAt`). `GetAlbumInLibrary` and `GetAlbumsInLibrary` populate the DTO. |
| `src/internal/library/adapters/dashboard.templ` | Album list row (`albumListRow`), `AlbumListRating`, `AlbumRowTagsSection`, `AlbumRating`. The row structure is where a notes indicator/button would live. |
| `src/internal/library/adapters/album_detail.templ` | Album detail page — renders rating, rating history, tags, tracks sections. Sleeve notes section would slot in here. |
| `src/internal/tags/adapters/http.go` | Best reference for the modal-trigger pattern: GET opens modal, POST closes it + OOB-swaps affected elements. |
| `src/internal/tags/adapters/tags.templ` | `TagsModal`, `CloseTagsModal` — canonical modal component pattern. |
| `src/internal/core/templates/icons.templ` | `NotesIcon` already exists — ready to use as the sleeve notes trigger. |
| `src/internal/server/server.go` | Route registration. New GET/PUT routes for notes would go here. |

## Architecture

**Current data flow for rating notes (not sleeve notes):**
The `note` column in `album_rating_log` is a per-rating-entry annotation. It is written alongside the rating score in `InsertAlbumRatingLogEntry` and displayed in `AlbumRatingHistory` on the detail page. It is tied to a specific rating event, not the album as a whole.

**What sleeve notes require:**
A separate, stable, user-editable text document per album. The cleanest approach is a new table — `album_notes` — with `(user_id, album_id)` as a unique key, holding a single `content text` column. Alternatively, a column could be added to a logical "user album metadata" table if one existed, but it does not.

**Data flow for sleeve notes (proposed):**
```
Album row / detail page → notes icon button (hx-get)
  → GET /app/notes/album?albumId=X
  → SleeveNotesModal(album) rendered inside modal overlay
  → User edits textarea
  → PUT /app/notes/album?albumId=X (hx-put)
  → notes.Service.UpsertNote()
  → DB upsert on album_notes
  → CloseNotesModal() + OOB-swap notes indicator in row/detail
```

**DTO chain:**
New `NoteDTO` in a `notes` package (or added to `review`) → embedded in `library.AlbumDTO.Note *NoteDTO` → passed to list row and detail page templates.

**Modal pattern** (established by tags):
- `GET` handler fetches album + existing note, renders modal component with `hx-swap="none"` (modal opens via JS).
- `PUT` handler validates, upserts, renders `CloseNotesModal()` + OOB-swap of indicator element.
- Modal component lives in `notes/adapters/` alongside the HTTP handler.

## Existing Patterns

- **New module:** `tags` is the best template. It has its own `service.go`, `adapters/http.go`, `adapters/tags.templ`, registered in `server.go` via a dedicated handler. A `notes` module following this structure would be consistent.
- **HTTP handlers:** Extract `userId` via `ctx.UserId()`, parse `albumId` from query, call service, render templ component. Use `httpx.HandleErrorResponse()` for all error paths.
- **Templ components:** Modal + form in `adapters/<feature>.templ`. Run `task build/templ` after edits. Import from `library/adapters` for OOB target components.
- **OOB swaps:** After save, render `CloseModal()` + any OOB components that need to update. Target elements need stable IDs (see `GetAlbumRatingID`, `GetAlbumRowTagsSectionID`).
- **DB queries:** New file in `db/queries/` (e.g., `album_notes.sql`). After editing, run `task build/sqlc`. Migrations via `task db/create -- migration_name` then `task db/up`.
- **Library DTO augmentation:** `library.AlbumDTO` is already enriched with ratings and tags fetched in `GetAlbumInLibrary` and `GetAlbumsInLibrary`. A `Note *NoteDTO` field follows the same pattern — bulk fetch for the list, single fetch for the detail.
- **Context:** Always use `contextx.ContextX`, never raw `context.Context`.

## Constraints & Risks

- **`album_rating_log.note` naming confusion:** The existing `note` field on rating log entries is a different concept from sleeve notes. Naming must be explicit to avoid confusion in code, queries, and UI (e.g., use "sleeve note" or "album note" in variable/method names rather than just `note`).
- **`AlbumDTO` enrichment cost:** `GetAlbumsInLibrary` already makes several DB round-trips (ratings, tags, listening history). Adding a bulk notes fetch adds one more. This is consistent with existing patterns but worth noting if performance becomes a concern.
- **SQLite `upsert`:** Use `INSERT OR REPLACE` or `INSERT ... ON CONFLICT(user_id, album_id) DO UPDATE SET content = EXCLUDED.content` — both are supported. The latter is preferred for explicitness.
- **Text length limit:** The existing `note` field in the rating flow enforces a 2000-character limit at the handler level. Sleeve notes may warrant a higher limit (they are meant for longer writing). The limit should be a named constant, enforced in the handler and optionally as a `CHECK` constraint.
- **Album detail page vs list row:** The detail page has more room and is the natural home for reading/editing sleeve notes. The list row should show only a small indicator icon (filled if notes exist, outline if not) — not the full text.
- **No new module required:** A `notes` package is clean but adds overhead. The feature could also live in the `review` module since it is conceptually part of a user's review of an album. This is an open question (see below).

## Open Questions

1. **Separate `notes` module or extend `review`?** The `review` module currently handles rating log entries and the rating recommender. Sleeve notes are logically related but structurally different (one stable record vs. append-only log). A dedicated `notes` module keeps concerns clean; adding to `review` avoids proliferating packages. Which is preferred?

2. **List row UI:** Should the list row show only an icon indicator (filled/outline based on whether notes exist), or also a short text preview (e.g., first line truncated in a tooltip)?

3. **Character limit:** What is the intended max length for sleeve notes? The rating `note` field caps at 2000 chars. Sleeve notes may want 5000 or more, or no limit.

4. **Rich text or plain text?** Is Markdown rendering in scope for v1, or is plain `<pre>`/`whitespace-pre-wrap` sufficient?

5. **Editing on the detail page inline (no modal) or always in a modal?** The detail page already has dedicated sections for rating, rating history, and tags. Sleeve notes could be an inline editable section on the detail page rather than a modal, which might feel more natural for longer text.

6. **Sortable/filterable by "has notes"?** Should the library list be filterable to show only albums with sleeve notes? This would touch `FilterParams` and the filter UI.

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research sleeve-notes -->
