# Rating Log — Implementation Plan

## Approach

Replace the mutable single-row `album_ratings` model with an append-only `album_rating_log` table. Every rating submission inserts a new row; the most recent row is the current rating. Individual log entries can be deleted from the Rating History section on the album detail page — any entry can be removed, not just the most recent. The rating modal no longer has a delete button. Notes are attached to a rating entry at submission time via a `note` column on the log table and are no longer separately editable.

The cleanest migration path is to create a new `album_rating_log` table, copy existing data into it, and drop the old `album_ratings` table. This avoids fighting SQLite's limited `ALTER TABLE` support and makes the semantic change explicit.

**Decisions made:**

- **Table name**: `album_rating_log` (not reusing `album_ratings`) — the append-only semantics are different enough that a new name is warranted.
- **`updated_at` column**: Dropped entirely. Each log entry has only `created_at`; there is no mutation after insert.
- **Notes on migrated rows**: Existing `review` values are copied into the new `note` column on the migrated log entry. Each historical row becomes its own log entry carrying the note that was on it.
- **`AlbumNotesIcon`**: Removed entirely from the dashboard. Notes functionality is no longer surfaced on the dashboard row — notes are only visible on the album detail page in the Rating History section.
- **Rating history rendering**: Rendered inline on the album detail page when the page loads (via an extended `GetAlbumInLibrary` or a new service method). A separate lazy-load route would add complexity without benefit for a small history list.
- **Deleting entries**: Any entry can be deleted from the Rating History section on the album detail page. Deletion is by entry ID. When no rows remain, the album is treated as unrated. The delete button is removed from the rating modal entirely.
- **Current rating after delete**: The most recent remaining entry becomes the current rating. If the deleted entry was the most recent, the rating badge on the detail page and dashboard updates accordingly.

---

## Files to Change

| File | Change |
|------|--------|
| `db/migrations/<timestamp>_album_rating_log.sql` | New migration: create `album_rating_log`, migrate data from `album_ratings`, drop `album_ratings` |
| `db/queries/album_ratings.sql` | Replace all queries with new ones targeting `album_rating_log`; add `GetUserAlbumRatingLog` (full history) |
| `src/internal/core/db/sqlc/models.go` | Regenerated — `AlbumRating` struct replaced by `AlbumRatingLog` |
| `src/internal/core/db/sqlc/album_ratings.sql.go` | Regenerated from updated queries |
| `src/internal/review/service.go` | `UpdateRating` becomes an INSERT; `ClearRating` becomes a DELETE of the most recent row; `UpdateReview` removed; add `GetRatingLog` |
| `src/internal/review/adapters/http.go` | Update `SubmitRatingRecommenderRating` to pass note from form; remove `DeleteRatingRecommenderRating` and `GetReviewNotes` and `SubmitReviewNotes` handlers; add `DeleteRatingLogEntry` handler |
| `src/internal/review/adapters/rating.templ` | Add `note` textarea to `RatingRecommenderConfirm`; pass note through confirm form |
| `src/internal/review/adapters/review_notes.templ` | Delete file (modal and form no longer needed) |
| `src/internal/library/service.go` | Update `GetAlbumsInLibrary` to use latest-per-album query; update `GetAlbumInLibrary` to also fetch rating log history; update `GetUnratedAlbums` query reference |
| `src/internal/library/adapters/album_detail.templ` | Replace "Notes" section with "Rating History" section showing all log entries, each with a delete button |
| `src/internal/library/adapters/dashboard.templ` | Remove `AlbumNotesIcon` component and the Notes menu item from the album row ellipsis dropdown; remove `GetAlbumNotesID` helper function |
| `src/internal/server/server.go` | Remove `GET /app/review/notes` and `POST /app/review/notes` route registrations |
| `e2e/feat/reviews.feature` | Update/replace scenarios to reflect new behaviour (note on confirm form, history on detail page, no separate notes modal) |

---

## Implementation Steps

1. **Create the migration** (`task db/create -- album_rating_log`). Write the Up block: create `album_rating_log(id, user_id, album_id, rating, note, created_at)` with no unique constraint; `INSERT INTO album_rating_log SELECT id, user_id, album_id, rating, review, created_at FROM album_ratings`; `DROP TABLE album_ratings`. Write the Down block as a no-op with an explanatory comment (SQLite cannot recreate the dropped table easily and data is gone). Apply with `task db/up`.

2. **Update SQL queries** in `db/queries/album_ratings.sql`. Remove `UpsertAlbumRating`, `UpsertAlbumReview`, `ClearAlbumRating`. Add:
   - `InsertAlbumRatingLogEntry :one` — simple INSERT RETURNING
   - `DeleteAlbumRatingLogEntry :exec` — DELETE WHERE id = ? AND user_id = ? (scoped to user for safety)
   - `GetLatestUserAlbumRating :one` — SELECT … ORDER BY created_at DESC LIMIT 1
   - `GetLatestUserAlbumRatings :many` — SELECT … WHERE id IN (subquery for max created_at per album per user) for the bulk library load
   - `GetUserAlbumRatingLog :many` — SELECT all entries for a user+album ORDER BY created_at DESC
   - Update `GetUnratedAlbums` to LEFT JOIN on `album_rating_log` and use the latest-entry subquery

   Regenerate with `task build/sqlc`.

3. **Update `review/service.go`**:
   - `AlbumRatingDTO`: rename `Review` field to `Note`; add `CreatedAt time.Time`
   - `NewAlbumRatingDTOFromModel`: update to map from new `AlbumRatingLog` model
   - `AddRating(ctx, userId, albumId string, rating float64, note string) (*AlbumRatingDTO, error)` — calls `InsertAlbumRatingLogEntry`
   - `DeleteRatingEntry(ctx, userId, entryId string) error` — calls `DeleteAlbumRatingLogEntry` by ID
   - Remove `UpdateReview`, `ClearRating`
   - Add `GetRatingLog(ctx, userId, albumId string) ([]*AlbumRatingDTO, error)` — calls `GetUserAlbumRatingLog`

4. **Update `library/service.go`**:
   - `GetAlbumsInLibrary`: replace `GetUserAlbumRatings` call with `GetLatestUserAlbumRatings`
   - `GetAlbumInLibrary`: replace `GetUserAlbumRating` with `GetLatestUserAlbumRating`; also call `reviewService.GetRatingLog` and attach to `AlbumDTO`
   - `AlbumDTO`: add `RatingLog []*review.AlbumRatingDTO` field

   Note: `library.Service` currently does not hold a `review.Service`. It calls DB queries directly. Either pass the review service into `library.Service`, or add `GetUserAlbumRatingLog` query directly and call it from the library service. The simpler path is adding the query call directly in the library service (consistent with how it calls rating queries today).

5. **Update `review/adapters/rating.templ`**:
   - Add a `note` textarea field to `RatingRecommenderConfirm` (below the rating input, above the submit button). Max 2000 chars. Optional.
   - Remove the trash/delete button from `RatingRecommenderConfirm` entirely.

   Run `task build/templ`.

6. **Update `review/adapters/http.go`**:
   - `SubmitRatingRecommenderRating`: parse `note` from form; pass to `reviewService.AddRating`
   - Remove `DeleteRatingRecommenderRating` — deletion moves to the detail page
   - Remove `GetReviewNotes` and `SubmitReviewNotes` entirely
   - Add `DeleteRatingLogEntry` handler: accepts entry `id` from the URL path, calls `reviewService.DeleteRatingEntry`, then re-renders the updated `AlbumRating` OOB swap and the updated rating history section

7. **Delete `review/adapters/review_notes.templ`** and its generated `_templ.go` file. Run `task build/templ` to confirm no orphaned references.

8. **Update `server/server.go`**: remove the two review notes route registrations and the `DELETE /app/review/rating-recommender/rating` route; add `DELETE /app/review/rating-log/{id}`.

9. **Update `library/adapters/album_detail.templ`**:
   - Replace the "Notes" section with a "Rating History" section listing all `album.RatingLog` entries in reverse-chronological order: date, score + label, note (if present), and a delete button per entry.
   - The delete button sends `DELETE /app/review/rating-log/{id}?albumId={albumId}`. On success the handler re-renders the rating badge (OOB swap) and the full history section in-place.
   - Remove the `AlbumNotesIcon` component from `dashboard.templ` and the Notes item from the album row ellipsis dropdown.

   Run `task build/templ`.

10. **Update `e2e/feat/reviews.feature`**: update scenarios to cover the new behaviour (note textarea on confirm form, no separate notes modal, rating history visible on detail page).

---

## Database Changes

### New migration

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE album_rating_log (
    id         text primary key,
    user_id    text not null references users(id) on delete cascade,
    album_id   text not null references albums(id) on delete cascade,
    rating     float not null,
    note       text,
    created_at datetime not null default current_timestamp
);

INSERT INTO album_rating_log (id, user_id, album_id, rating, note, created_at)
SELECT id, user_id, album_id, rating, review, created_at
FROM album_ratings
WHERE rating IS NOT NULL;

DROP TABLE album_ratings;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'album_rating_log migration cannot be rolled back; data from album_ratings has been dropped';
-- +goose StatementEnd
```

Note: only rows where `rating IS NOT NULL` are migrated — existing rows that only had a review (null rating) would not be meaningful as a rating log entry.

### Query changes summary

- `UpsertAlbumRating` → `InsertAlbumRatingLogEntry`
- `UpsertAlbumReview` → removed
- `ClearAlbumRating` → `DeleteAlbumRatingLogEntry` (by entry ID, not by most-recent)
- `GetUserAlbumRatings` → `GetLatestUserAlbumRatings` (uses subquery for latest-per-album)
- `GetUserAlbumRating` → `GetLatestUserAlbumRating`
- New: `GetUserAlbumRatingLog`
- `GetUnratedAlbums` → updated JOIN and filter condition

---

## Feature Specs

```gherkin
Feature: Rating Log

  Users can submit multiple ratings for an album over time. Each submission
  creates a new entry in the rating log. The most recent entry is the current
  rating. The album detail page shows the full rating history.

  Scenario: Submitting a rating with an optional note
    Given a logged-in user with an album in their library
    When they open the rating modal, enter a score, optionally add a note, and click Lock in
    Then the modal closes and the album shows the new rating

  Scenario: Rating and note are recorded together
    Given a logged-in user who submits a rating with a note
    When they view the album detail page
    Then the rating history shows the new entry with its score and note

  Scenario: Submitting a second rating creates a new history entry
    Given a logged-in user who has already rated an album
    When they submit a new rating for the same album
    Then the album detail page shows two entries in the rating history
    And the most recent entry is displayed as the current rating

  Scenario: Deleting a non-current rating entry from history
    Given a logged-in user who has submitted two ratings for an album
    When they navigate to the album detail page and click delete on the older entry
    Then that entry is removed from the history
    And the current rating is unchanged

  Scenario: Deleting the current rating from history rolls back to the previous one
    Given a logged-in user who has submitted two ratings for an album
    When they navigate to the album detail page and click delete on the most recent entry
    Then that entry is removed from the history
    And the previous rating is now shown as the current rating

  Scenario: Deleting the only rating entry clears the album rating
    Given a logged-in user who has rated an album exactly once
    When they navigate to the album detail page and click delete on the only entry
    Then the history section is empty
    And the album shows no rating

  Scenario: No delete button in the rating modal
    Given a logged-in user opening the rating modal for a rated album
    When the confirm form is shown
    Then no delete button is visible

  Scenario: Rating history is shown on the album detail page
    Given a logged-in user who has submitted multiple ratings for an album
    When they navigate to the album detail page
    Then a Rating History section is visible listing all past entries in reverse-chronological order

  Scenario: No notes on dashboard
    Given a logged-in user viewing the library dashboard
    When they look at an album row
    Then no notes icon or notes editing button is present
```

---

## Testing

**Unit / integration (Go):**
- `review.Service.UpdateRating`: verify a new row is inserted on each call, not upserted.
- `review.Service.ClearRating`: verify only the most-recent row is deleted.
- `review.Service.GetRatingLog`: verify entries returned in reverse-chronological order.
- `library.Service.GetAlbumsInLibrary`: verify the most-recent rating per album is selected, not an arbitrary one.
- `GetUnratedAlbums` query: verify albums with a log entry are excluded, albums without any entry are included.

**E2E (`e2e/feat/reviews.feature`):**
- Cover all Gherkin scenarios above.
- Verify the note textarea is present on the rating confirm form.
- Verify the Rating History section on the album detail page shows multiple entries after submitting multiple ratings.
- Verify delete rolls back to previous rating.
- Verify the old notes modal route (`/app/review/notes`) no longer exists (returns 404 or is unreachable via the UI).

---

## Risks & Mitigations

**Migration is destructive**: The `album_ratings` table is dropped. Make sure the migration is tested locally before applying to any shared environment. The Down block intentionally does nothing — this is by design (append-only is a one-way conversion).

**`GetUserAlbumRatings` fan-out query**: The new `GetLatestUserAlbumRatings` requires a subquery to pick the latest entry per album. In SQLite this is straightforward with `WHERE (user_id, album_id, created_at) IN (SELECT user_id, album_id, MAX(created_at) FROM album_rating_log WHERE user_id=? GROUP BY album_id)`. Test with a user who has multiple entries per album to confirm only the latest is returned.

**`AlbumDTO.RatingLog` in library service**: Adding a new field and a new DB call in `GetAlbumInLibrary` increases queries per page load by one. This is acceptable for the single-album detail page. It should not be added to `GetAlbumsInLibrary` (bulk load for the dashboard) — history is only needed on the detail page.

**Template compilation after deleting `review_notes.templ`**: Any import or reference to `ReviewNotesModal`, `CloseReviewNotesModal`, or `ReviewNotesForm` will cause a compile error. Grep for these symbols before deleting the file and remove all call sites first.

**`AlbumNotesIcon` removal**: The `AlbumNotesIcon` component in `dashboard.templ` and the Notes item in the album row ellipsis dropdown must be fully deleted. Grep for `GetAlbumNotesID`, `AlbumNotesIcon`, `album-row-notes`, and `/app/review/notes` references in templates before deleting to avoid broken references.

---

## Feedback
<!-- Review this document and add your feedback here, then re-run /feature-plan rating-log -->
