# Rating Log — Research

## Summary

The rating-log feature converts the current single-row `album_ratings` model into an append-only log. Instead of one mutable rating per (user, album) pair, every new rating submission creates a new entry in a history table. The most recent entry is shown as the current rating everywhere in the UI; the album detail page gains a **Rating History** section listing all past entries in reverse-chronological order. Notes are tied to their specific rating entry at submission time and are not separately editable afterward. The trash button in the rating modal deletes only the most recent entry — rolling back to the previous one if it exists.

---

## Relevant Code

### Domain & Service Layer

- `/src/internal/review/service.go` — `Service` with `UpdateRating`, `ClearRating`, `UpdateReview`. Currently issues an UPSERT to a single `album_ratings` row. This is the primary file to change.
- `/src/internal/review/rating.go` — `RatingQuestions`, `RatingKey`, `GetRatingLabel`, all rating computation logic. No changes needed here.
- `/src/internal/review/adapters/http.go` — HTTP handlers: `SubmitRatingRecommenderRating`, `DeleteRatingRecommenderRating`, `SubmitReviewNotes`. The submit handlers call `UpdateRating`/`UpdateReview`; the delete handler calls `ClearRating`. All need revisiting.

### Templates

- `/src/internal/review/adapters/rating.templ` — `RatingModal`, `RatingRecommenderConfirm`, `RatingRecommenderQuestions`. The confirm form currently shows an optional note textarea should be added here (per wiki spec: note is attached at submission time).
- `/src/internal/review/adapters/review_notes.templ` — `ReviewNotesForm` and `ReviewNotesModal`. Per the new spec, the note is no longer a separately-editable field; this modal's role needs reconsidering. It may be replaced by the note input on the confirm form.
- `/src/internal/library/adapters/album_detail.templ` — Album detail page. Needs a new **Rating History** section after the current rating display.
- `/src/internal/library/adapters/dashboard.templ` — `AlbumRating` component (badge or button in table rows). No behavioral change; the most recent rating entry continues to be displayed here.

### Data Layer

- `/db/migrations/20260303132728_album_ratings.sql` — Creates the `album_ratings` table with `unique(user_id, album_id)` — this unique constraint is the core thing to remove/replace.
- `/db/migrations/20260306142605_album_ratings_review.sql` — Adds `review text` column to `album_ratings`.
- `/db/queries/album_ratings.sql` — Contains `UpsertAlbumRating`, `UpsertAlbumReview`, `GetUserAlbumRatings`, `GetUserAlbumRating`, `ClearAlbumRating`, `GetUnratedAlbums`. The upsert queries and clear query all need replacing with insert/delete semantics.
- `/src/internal/core/db/sqlc/models.go` — `AlbumRating` struct with `ID`, `UserID`, `AlbumID`, `Rating`, `CreatedAt`, `UpdatedAt`, `Review`. Will need to be regenerated after schema changes.
- `/src/internal/core/db/sqlc/album_ratings.sql.go` — Generated from `album_ratings.sql`; regenerated via `task build/sqlc`.

### Library Integration

- `/src/internal/library/service.go` — `GetAlbumsInLibrary` and `GetAlbumInLibrary` call `GetUserAlbumRatings`/`GetUserAlbumRating` and map to `AlbumRatingDTO`. After the schema change, these queries must return only the most recent entry per (user, album).
- `AlbumDTO.Rating *review.AlbumRatingDTO` — The DTO holds a pointer to the current rating. This structure works unchanged if the query returns the most recent entry.

### Routing

- `/src/internal/server/server.go` — Review routes are: `GET /app/review/rating-recommender`, `GET /app/review/rating-recommender/questions`, `POST /app/review/rating-recommender/questions`, `POST /app/review/rating-recommender/rating`, `DELETE /app/review/rating-recommender/rating`, `GET /app/review/notes`, `POST /app/review/notes`. A new route for fetching rating history will be needed for the album detail page (or it can be inlined when the detail page loads).

---

## Architecture

### Current Flow

```
User clicks "Rate" on album row or detail page
  → GET /app/review/rating-recommender?albumId=<id>
  → Returns RatingModal (opens as modal overlay)
    → Either questionnaire → confirm form, or confirm form directly
  → POST /app/review/rating-recommender/rating?albumId=<id>
    → reviewService.UpdateRating() → UpsertAlbumRating (single row per user+album)
    → Closes modal, OOB-swaps AlbumRating component
  → DELETE /app/review/rating-recommender/rating?albumId=<id>
    → reviewService.ClearRating() → NULL-sets rating field on single row

Notes are separate:
  → GET /app/review/notes?albumId=<id>
  → POST /app/review/notes → UpdateReview() → UpsertAlbumReview (updates review on same row)
```

### Target Flow (append-only)

```
Submit rating + optional note:
  → INSERT new row into album_rating_log (user_id, album_id, rating, note, created_at)
  → Current rating = most recent row by created_at

Delete (trash):
  → DELETE the most recent row for (user_id, album_id) by created_at
  → Previous row, if any, becomes current

Display current rating:
  → SELECT ... WHERE user_id=? AND album_id=? ORDER BY created_at DESC LIMIT 1

Display history (album detail page):
  → SELECT ... WHERE user_id=? AND album_id=? ORDER BY created_at DESC
```

Notes are no longer a separate editable field — they are attached to a rating entry at creation time and become immutable. The `review_notes` modal flow collapses into the rating confirm form.

---

## Existing Patterns

- **Migrations**: Use goose format with `-- +goose Up` / `-- +goose Down` blocks. Create with `task db/create -- <name>`. Apply with `task db/up`. SQLite does not support `DROP COLUMN` in rollbacks.
- **SQLC queries**: Add named queries to `.sql` files in `db/queries/`, regenerate with `task build/sqlc`. Generated code lands in `/src/internal/core/db/sqlc/`.
- **Service layer**: Business logic isolated in `Service` structs. HTTP handlers call service methods. No direct DB access from handlers.
- **Context**: Use `contextx.ContextX`, extract user ID via `ctx.UserId()`.
- **Error handling**: `httpx.HandleErrorResponse()` for all HTTP-layer errors.
- **HTMX OOB swaps**: After mutating actions, handlers render `CloseModal` + updated component with `hx-swap-oob="true"` to update the relevant table cell or detail page section without a full reload. See `DeleteRatingRecommenderRating` for the pattern.
- **Templ**: `.templ` files compiled with `task build/templ`. Generated `_templ.go` files are committed.
- **Modal pattern**: Modals are rendered server-side and injected into the DOM via `hx-swap="none"` on trigger elements; a dedicated close component (`CloseRatingModal`, `CloseReviewNotesModal`) forces the modal closed.

---

## Constraints & Risks

### Schema Migration Complexity

The current `album_ratings` table has a `unique(user_id, album_id)` constraint that enforces one row per user-album pair — the opposite of the append-only model. Options:

1. **Rename and recreate**: Create a new `album_rating_log` table, migrate existing data, drop or archive `album_ratings`. SQLite's lack of `ALTER TABLE ... DROP CONSTRAINT` makes in-place changes messy.
2. **Add a new table, keep old**: Add `album_rating_log` as the new truth for new entries while keeping `album_ratings` for backward compatibility. More migration complexity, messy dual-source.

Option 1 is cleaner. The migration must copy existing `(id, user_id, album_id, rating, review, created_at)` rows into the new table, then drop the old table.

### Review Notes Coupling

`review_notes.templ` and the `POST /app/review/notes` route currently manage notes as a separate, re-editable field. Per the new spec, notes attach to a rating entry and cannot be edited separately. This means:
- The review notes modal and its route (`GET /app/review/notes`, `POST /app/review/notes`) are likely removed or repurposed.
- The confirm form (`RatingRecommenderConfirm`) gains a note textarea.
- The `AlbumNotesIcon` component (shown on dashboard rows and album detail) needs rethinking — if notes are per-entry rather than per-album, there is no single "album has a note" state; only the most recent entry may or may not have a note.

### "GetUserAlbumRatings" Query Fan-out

`GetAlbumsInLibrary` in `library/service.go` calls `GetUserAlbumRatings` to fetch all ratings for a user at once, then maps them by album ID. This query must return only the most recent entry per album (e.g. `SELECT ... WHERE id IN (SELECT MAX(id)/created_at subquery)`). The current simple `select * from album_ratings where user_id = ?` will not work as-is.

### `GetUnratedAlbums` Query

This query joins `album_ratings` and tests `album_ratings.rating IS NULL`. After the schema change it needs updating to check whether a current (most-recent) rating log entry exists and has a non-null score.

### AlbumRatingDTO Shape

`AlbumRatingDTO` currently holds `ID`, `UserID`, `AlbumID`, `Rating *float64`, `Review *string`. In the new model, "the current rating's note" is the note on the most recent log entry. The DTO can largely stay the same if queries return only the most recent entry, but `UpdatedAt` semantics change — each entry has its own `created_at` and there is no `updated_at`.

---

## Open Questions

1. **Table rename strategy**: Should the new table be `album_rating_log` (new name) or keep `album_ratings` after dropping the unique constraint? Keeping the name reduces diff but may be confusing given the semantic change.

2. **Notes on existing ratings**: After migration, existing rows in `album_ratings` have a `review` field. Does that review become the note on the migrated log entry? If so, it should be copied into the new `note` column.

3. **AlbumNotesIcon on the dashboard row**: If notes are per-entry (not per-album), should the notes icon on dashboard rows reflect whether the most recent rating entry has a note, or disappear entirely? The wiki spec says "notes are tied to the specific rating entry" — does that mean the icon goes away from the row, or does it show if the most recent entry has a note?

4. **Rating history on album detail — inline or separate route?**: The wiki says the album detail page shows a Rating History section. Should this be rendered inline when the page loads (using existing `GetAlbumInLibrary` extended to return history), or fetched lazily via a new route (e.g. `GET /app/review/rating-history?albumId=<id>`)?

5. **Deleting the only remaining entry**: The trash button deletes the most recent entry. If it's the last entry, the rating is cleared entirely. Should the handler also delete the log row (leaving no row) rather than nulling a field? This is cleaner with an append-only log — deletion = removing the row.

6. **`UpdatedAt` column**: The current `album_ratings` schema has `updated_at`. In an append-only log, there is no update — should the new table drop `updated_at` entirely and rely solely on `created_at`?

---

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-research rating-log -->
