# Album Text Review/Notes — Research

## Summary

Allow users to add and update free-text notes/reviews on each album in their library. This is a v1 item in `docs/feat/ranking-review.md`. The numeric rating system is already complete; this feature adds an optional text field alongside it, stored in the same `album_ratings` row (one row per user/album).

## Relevant Code

| Path | Role |
|------|------|
| `db/migrations/20260303132728_album_ratings.sql` | Schema for `album_ratings` table |
| `db/queries/album_ratings.sql` | SQL queries — upsert, get by user/album |
| `src/internal/core/db/sqlc/album_ratings.sql.go` | Generated query code (do not edit) |
| `src/internal/core/db/sqlc/models.go` | Generated `AlbumRating` struct |
| `src/internal/review/service.go` | `Service`, `AlbumRatingDTO`, `UpdateRating()` |
| `src/internal/review/adapters/http.go` | HTTP handlers for rating flow |
| `src/internal/review/adapters/rating.templ` | Rating modal templates |
| `src/internal/library/service.go` | `AlbumDTO` (holds `Rating *AlbumRatingDTO`), `GetAlbumInLibrary`, `GetAlbumsInLibrary` |
| `src/internal/library/adapters/dashboard.templ` | `AlbumRating` badge, `albumRow`, `AlbumsTable` |
| `src/internal/server/server.go` | Route registration |

## Architecture

**Data flow for ratings (the existing pattern to mirror):**

```
Album row click → AlbumRating badge (hx-get)
  → GET /app/review/rating-recommender?albumId=X
  → RatingModal (templ) inside modal overlay
  → User fills questions or direct rating
  → POST /app/review/rating-recommender/rating?albumId=X
  → review.Service.UpdateRating()
  → DB upsert on album_ratings
  → CloseRatingModal() + OOB-swap AlbumRating badge
```

**DB:** `album_ratings` has a unique constraint on `(user_id, album_id)`, so one row per user/album. The upsert uses `COALESCE(EXCLUDED.rating, rating)` to preserve existing values when not provided.

**DTO chain:** `sqlc.AlbumRating` → `review.NewAlbumRatingDTOFromModel()` → `review.AlbumRatingDTO` → embedded in `library.AlbumDTO.Rating` → passed to all templates.

**Library service** fetches ratings in bulk (`GetUserAlbumRatings`) for the dashboard and individually (`GetUserAlbumRating`) for single-album views. Both paths populate `AlbumDTO.Rating`.

## Existing Patterns

- **Service methods:** `func (s *Service) UpdateRating(ctx context.Context, userId, albumId string, rating float64) (*AlbumRatingDTO, error)` — upserts via sqlc, returns DTO.
- **HTTP handlers:** extract `userId` via `ctx.UserId()`, parse query/form params, call service, render templ component. Use `httpx.HandleErrorResponse()` for errors.
- **Templ components:** live in `adapters/` alongside the HTTP handler. Run `templ generate` after editing `.templ` files.
- **HTMX:** forms use `hx-post`/`hx-put`; responses are HTML fragments. OOB swaps (`hx-swap-oob="true"`) update other parts of the page after a save.
- **SQL changes:** edit `db/queries/`, run `sqlc generate`, commit both the `.sql` and generated `.go` files.
- **Migrations:** `goose -dir db/migrations create <name> sql` then `task db/up`.

## Constraints & Risks

- **Upsert COALESCE logic:** The current upsert preserves `rating` when not provided. A new `review` column must follow the same pattern — `COALESCE(EXCLUDED.review, review)` — so saving a rating doesn't wipe the review and vice versa. This is straightforward but easy to miss.
- **`album_ratings` is shared state:** `AlbumRatingDTO` is returned from both bulk and single-album fetches. Adding `Review *string` to the DTO affects both paths with no extra work.
- **`sqlc generate`:** After updating the SQL query, regenerated Go files must be committed. The `models.go` `AlbumRating` struct will gain a `Review sql.NullString` field automatically.
- **No XSS risk in templ:** Templ auto-escapes string interpolation, so raw text review content is safe to render directly.

## Open Questions

1. **Entry point in the UI:** Should the review be editable from:
   - A separate icon/button in the album row (parallel to the rating badge)?
   - A tab or additional step inside the existing rating modal?
   - Both (rating badge already opens the modal; a notes icon could open a separate one)?

2. **Independent of rating?** Can a user save notes without having a numeric rating? The DB allows it (`rating` is nullable), but should the UI enforce a rating first?

3. **Display in the album table:** Show a notes indicator icon when a review exists, or show a short text preview/tooltip?

4. **Length limit:** Should review text be capped (e.g., 2000 chars)? If so, enforce at DB level (CHECK constraint), HTTP handler, or just frontend?

5. **Timestamped history:** The feature doc mentions journal-style entries with timestamps. Is this v1 (meaning one editable text field with `updated_at`) or v1.1 (append-only log)?

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research album-review-notes -->
