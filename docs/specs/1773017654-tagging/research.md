# Tagging — Research

## Summary

Allow users to apply custom tags to albums in their library, organized by named tag groups. The v1 scope (from `docs/feat/tagging.md`) covers: unlimited tags per album, two predefined tag groups ("Sound" and "Mood"), support for ungrouped tags, and no cap on tags per group. Tags are user-scoped — each user manages their own tag taxonomy. This research examines where the feature plugs into the existing data layer, UI patterns, and service architecture.

## Relevant Code

| Path | Role |
|------|------|
| `db/schema.sql` | Authoritative current schema (no tagging tables yet) |
| `db/migrations/` | Goose migration files; new tables require a migration here |
| `db/queries/album_ratings.sql` | Pattern reference: upsert queries for user-scoped data |
| `src/internal/core/db/sqlc/` | Generated sqlc output — do not edit directly |
| `src/internal/core/db/sqlc/models.go` | Generated model structs; custom typed enums live in `db/models/` |
| `src/internal/core/db/models/models.go` | Hand-written domain enums/types used in sqlc overrides |
| `src/internal/review/service.go` | Pattern reference: `Service` struct with `db *db.DB`, DTO construction |
| `src/internal/review/adapters/http.go` | Pattern reference: handler → service → OOB-swap templ response |
| `src/internal/review/adapters/review_notes.templ` | Pattern reference: modal with a simple form |
| `src/internal/library/service.go` | `AlbumDTO` — where tag data would be embedded |
| `src/internal/library/adapters/dashboard.templ` | `albumRow`, `AlbumsTable` — where tag display/entry points live |
| `src/internal/server/server.go` | Route registration and service wiring |
| `src/internal/core/contextx/contextx.go` | `ctx.UserId()` for extracting the authenticated user |
| `src/internal/core/httpx/handler.go` | `httpx.HandleErrorResponse()` for error responses |
| `src/internal/core/templates/modal.templ` | Reusable `Modal`/`ForceCloseModal` pattern |
| `sqlc.yaml` | sqlc config — overrides needed for any new typed columns |

## Architecture

### Data Model (proposed)

The feature requires three new tables:

```
tag_groups
  id          text PK
  user_id     text FK → users(id)
  name        text           -- "Sound", "Mood", or user-defined
  created_at  datetime

tags
  id          text PK
  user_id     text FK → users(id)
  name        text
  group_id    text NULL FK → tag_groups(id)  -- NULL = ungrouped
  created_at  datetime
  UNIQUE(user_id, name)                       -- tags are global per user, not per-album

album_tags
  id          text PK
  user_id     text FK → users(id)
  album_id    text FK → albums(id)
  tag_id      text FK → tags(id)
  created_at  datetime
  UNIQUE(user_id, album_id, tag_id)
```

Tags are a shared resource owned by a user (not album-specific). A user creates "rainy day" once and can apply it to many albums. The `album_tags` join table links albums to tags within a user's library.

### Data Flow (mirroring existing patterns)

```
Album row action (tag icon or album detail)
  → GET /app/tags/album?albumId=X
  → TagsModal (templ) via OOB swap into #global-modal-container
  → User types/selects tags, submits
  → POST /app/tags/album?albumId=X
  → tags.Service methods (create tag if new, create album_tag)
  → CloseTagsModal() + OOB-swap updated tag display in album row
```

### Module Layout

A new `tags` module would mirror `review`:

```
src/internal/tags/
  service.go           -- Service struct, DTOs, business logic
  adapters/
    http.go            -- HTTP handlers
    tags.templ         -- Tag UI components
    tags_templ.go      -- Generated
```

### DTO Chain

`sqlc` rows → `tags.NewTagDTOFromModel()` → `tags.TagDTO` (with group name) → embedded as `Tags []TagDTO` in `library.AlbumDTO` → rendered in templates.

The `library.Service.GetAlbumsInLibrary()` bulk-fetch loop already makes N queries for albums, artists, tracks, ratings — tags would be fetched similarly using a batch query by album IDs.

## Existing Patterns

- **Service struct:** `type Service struct { db *db.DB }` — same pattern used by `review.Service` and `library.Service`.
- **DTO construction:** `NewXxxDTOFromModel(model sqlc.Xxx) XxxDTO` — convert sqlc struct to domain DTO in the service layer.
- **HTTP handlers:** `ctx := contextx.NewContextX(r.Context())`, then `ctx.UserId()`. Error handling via `httpx.HandleErrorResponse()` with `httpx.HandleErrorResponseProps{Status: ..., Err: ...}`.
- **Templ modals:** `templates.Modal(id, templates.ModalProps{ModalContent: ...})` renders into `#global-modal-container` via OOB swap. `templates.ForceCloseModal(id)` closes it.
- **OOB swap after mutation:** After saving, handlers render `CloseXxxModal()` then the updated in-row component with `hx-swap-oob="true"` — see `SubmitReviewNotes` and `SubmitRatingRecommenderRating`.
- **SQL upserts:** Queries use `INSERT ... ON CONFLICT DO UPDATE` via sqlc. Generated code is in `src/internal/core/db/sqlc/`.
- **Migrations:** `task db/create -- migration_name` creates a goose file; `task db/up` runs it, dumps `db/schema.sql`, and re-runs `task build/sqlc`.
- **Typed enums:** Custom types go in `src/internal/core/db/models/models.go` and are registered in `sqlc.yaml` overrides.
- **AlbumDTO extension:** Adding `Tags []TagDTO` to `library.AlbumDTO` follows the same pattern as `Rating *review.AlbumRatingDTO` — both are fetched in bulk by the library service and passed through to templates.

## Constraints & Risks

- **Bulk fetch performance:** `GetAlbumsInLibrary` already joins albums, artists, tracks, and ratings. Adding tags means another round-trip or a larger join. A query like `GetAlbumTagsByAlbumIds(ctx, albumIds)` (matching the existing `GetAlbumArtistsByAlbumIds` pattern) keeps it consistent and avoids N+1.
- **Tag uniqueness per user:** The `UNIQUE(user_id, name)` constraint on `tags` means "rainy day" is one row — applying it to a second album is just inserting a new `album_tags` row, not a new tag. The service needs to handle `get-or-create` for tags by `(user_id, name)`.
- **Ungrouped tags:** `group_id NULL` handles this naturally, but queries that group tags by their group need to handle NULL carefully (LEFT JOIN or COALESCE).
- **Alpine.js / HTMX tag input:** A multi-tag input (type to search existing tags, create new on enter) is more complex than the existing radio/text field forms. This is the main frontend challenge. The existing codebase uses Alpine.js for reactive state (see `ratingConfirmAlpineData` in `rating.templ`) — a tag autocomplete component would likely need Alpine.js too.
- **Deleting tags vs. album_tags:** Removing a tag from an album deletes an `album_tags` row. Deleting the tag entirely (from the user's vocabulary) should cascade-delete all `album_tags` referencing it — the FK `ON DELETE CASCADE` must be set correctly.
- **Predefined groups bootstrap:** "Sound" and "Mood" groups are specified in v1. These are user-owned rows, not system-level constants, so they must be created for each user (either on first login, on first tag action, or seeded during library creation).

## Open Questions

1. **Where in the album row does tagging live?** The album table row currently has: album title, artists, rating badge, format icons, date added, last played, and an ellipsis dropdown (with "Notes"). Does tagging get a dedicated column (tag chips inline), an entry in the dropdown menu, or a separate icon like the notes icon?

2. **Tag display in the library table:** Should applied tags show as chips/badges in the album row, or only be visible when you open the tag editor? Inline chips could get noisy with many tags.

3. **Are predefined "Sound" and "Mood" groups created automatically?** If so, when — on first sync, on first tag action, or on account creation? They are user-owned rows and need to be seeded somehow.

4. **Tag filtering / search:** The feature doc mentions "intuitive search experience." Is filtering the albums table by tag in scope for v1, or is that a follow-on? This affects whether tags need to be included in the `GetAlbumsTable` query path.

5. **Tag input UX:** Is a plain text comma-separated input acceptable for v1, or does v1 require a proper tag autocomplete/combobox (show existing tags, create new on enter)?

6. **Shared vs. per-user tags:** The design assumes tags are fully user-scoped. Should there ever be a system-level or shared tag vocabulary (e.g., for genre suggestions), or is the taxonomy always 100% personal?

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research tagging -->
