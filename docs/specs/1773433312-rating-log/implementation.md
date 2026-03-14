# Rating Log — Implementation

## What was built

- New `album_rating_log` table (append-only, no unique constraint) replacing `album_ratings`
- Migration copies existing rated rows (rating IS NOT NULL) into the new table and drops the old one
- New SQL queries: `InsertAlbumRatingLogEntry`, `DeleteAlbumRatingLogEntry`, `GetLatestUserAlbumRating`, `GetLatestUserAlbumRatings`, `GetUserAlbumRatingLog`, updated `GetUnratedAlbums`
- `review.Service`: `AddRating`, `DeleteRatingEntry`, `GetRatingLog` replace `UpdateRating`, `ClearRating`, `UpdateReview`
- `AlbumRatingDTO.Review` renamed to `Note`; `CreatedAt` added; `Rating` is now `*float64` pointing to model value (always non-nil on real entries)
- `library.AlbumDTO.RatingLog []*review.AlbumRatingDTO` field added; `GetAlbumInLibrary` populates it
- `RatingRecommenderConfirm`: delete button removed; optional note textarea added
- `review_notes.templ` and its generated file deleted; `GetReviewNotes` and `SubmitReviewNotes` handlers removed
- Album detail page: Notes section replaced with Rating History section listing all log entries with per-entry delete buttons
- New `AlbumRatingHistory` component for OOB swap on entry delete
- `AlbumNotesIcon` component and Notes menu item removed from dashboard
- Routes: removed `DELETE /app/review/rating-recommender/rating`, `GET /app/review/notes`, `POST /app/review/notes`; added `DELETE /app/review/rating-log/{id}`
- E2E tests updated: old notes and delete-in-modal scenarios removed; new scenarios added for note textarea, rating history, delete from history

## Differences from the plan

- **`SubmitRatingRecommenderRating` fetches album after inserting rating**: The plan said to pass `albumRating` to `album.Rating = albumRating`, but the album is re-fetched from the library service after inserting so that `RatingLog` is also up to date for OOB swaps. The re-fetch is done inside `SubmitRatingRecommenderRating` after `AddRating`.
- **`GetLatestUserAlbumRatings` params are `UserID` and `UserID_2`**: SQLC generated duplicate param names because the same user_id appears in both the subquery and outer query. The library service passes `userId` for both fields.
- **`GetUnratedAlbums` params are `UserID`, `UserID_2`, `UserID_3`**: Same SQLC disambiguation for the three user_id occurrences in the query.
- **`AlbumRatingHistory` added as separate exported component**: The plan described OOB-swapping the history section but didn't specify the component structure. `AlbumRatingHistory(album, isOobSwap bool)` was extracted so the delete handler can render it with `hx-swap-oob="true"`.

## Plan inaccuracies

- **`review.Service` does not import `sqlx`**: The plan referred to `sqlx.NewNullString` but the package is `database/sql` — `sql.NullString{String: note, Valid: true}` is used directly.
- **`NewAlbumRatingDTOFromModel` always sets `Rating` (non-nil)**: In the new schema `rating float not null`, so `model.Rating` is a plain `float64`. The DTO sets `dto.Rating = &model.Rating` unconditionally. The plan described it the same way, but the old code used `sql.NullFloat64`.
- **`library.Service` does not take a `review.Service` param**: The plan discussed either passing `review.Service` into `library.Service` or calling DB queries directly. Queries are called directly (consistent with existing pattern), no constructor change needed.
- **`GetAlbumInLibrary` error check order**: The existing code used `errors.Is(sql.ErrNoRows, err)` (args reversed). The replacement uses `errors.Is(err, sql.ErrNoRows)` (correct order).
