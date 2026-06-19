# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Single topic file `review.go` holds the `RatingState` enum, `RatingStateDTO` (per-album lifecycle marker), `AlbumRatingDTO` (rating log entry), the scoring questionnaire, and the rating labels. `RatingState` is shared between the live state machine and the log entry, so they live together.
- Live `RatingState` is the two-value set `{provisional, finalized}`. `unrated` is the absence of a row in `album_rating_state` — there is no third stored value. The constants in `review.go` (`RatingStateProvisional`, `RatingStateFinalized`) are the only values the live system produces.
- The only application-driven transitions are `(no row) → provisional` (implicit on first save), `provisional → finalized` (explicit, via Finalize), and self-edges `provisional → provisional` / `finalized → finalized` on re-rate. The save path never transitions state; only `FinalizeWithRating` does.
- `FinalizeWithRating` is the manual finalize entry point — it writes a new rating-log entry with the supplied score and promotes the row to `finalized`. It returns `ErrFinalizeRequiresProvisional` (and writes nothing) when the album is unrated or already finalized.
- The modal entry route `GET /app/review/rating-recommender` always renders the score-entry form (`RatingConfirmFormFrag`), pre-filled from the most-recent rating-log score. The questionnaire is opt-in via a button on that form, carries `priorRating` so dismissal can restore the pre-fill, and never writes a rating row itself. The Finalize submit button on the score-entry form is gated on `RatingState.State == provisional`.
- `album_rating_log.state`'s CHECK still admits `'stalled'` because history is immutable — `RatingStateLogLabel` provides the display label for historical entries. The live `album_rating_state.state` CHECK is narrowed to `{provisional, finalized}` after the `20260517000001_retire_rerate_machinery.sql` migration.
- After saving/finalizing a rating or deleting a rating-log entry, handlers broadcast the `album-changed` HTMX event (via `httpx.SetHXTrigger`, detail `{"albumId": <id>}`) instead of rendering library views; library owns the refresh via its `GET /app/library/album-surfaces` endpoint.
