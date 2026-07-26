# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Single topic file `review.go` holds the `RatingState` enum, `RatingStateDTO` (per-album lifecycle marker), `AlbumRatingDTO` (rating log entry), the scoring questionnaire, and the rating labels. `RatingState` is shared between the live state machine and the log entry, so they live together.
- Live state machine entry points: `SaveRating` (always lands `provisional`) and `FinalizeWithRating` (always lands `finalized`). The constants `RatingStateProvisional` / `RatingStateFinalized` are the only values the live system emits; `unrated` is the absence of a state row. Full lifecycle spec and transitions: see domain doc below.
- `FinalizeWithRating` is the manual finalize entry point — it writes a new rating-log entry with the supplied score and upserts the state row to `finalized`.
- The modal entry route `GET /app/review/rating-recommender` always renders the score-entry form (`RatingConfirmFormFrag`), pre-filled from the most-recent rating-log score. Both save buttons render unconditionally — **Save & finalize** (primary) posts to the finalize route, **Save only** (secondary, the form's own submit) posts to the save route. The questionnaire is opt-in via a **Help me score it** link beneath the rating input, carries `priorRating` so dismissal can restore the pre-fill, and never writes a rating row itself.
- `album_rating_log.state`'s CHECK still admits `'stalled'` because history is immutable; `RatingStateLogLabel` provides the display label for historical entries. The live `album_rating_state.state` CHECK is narrowed to `{provisional, finalized}` after `20260517000001_retire_rerate_machinery.sql`.
- After saving/finalizing or deleting a rating-log entry, handlers broadcast `album-changed` via `httpx.SetHXTrigger` (detail `{"albumId": <id>}`); library owns the refresh via `GET /app/library/album-surfaces`.

## Domain docs

| Doc | What | When |
|---|---|---|
| [review](../../../docs/domain/review.md) | rating lifecycle state machine, questionnaire weights, rating log spec | working on rating reads/writes or lifecycle state; understanding score derivation or log behaviour |

## Product docs

| Doc | What | When |
|---|---|---|
| [ratings](../../../docs/product/ratings.md) | score albums 0–10 with questionnaire; provisional/finalized lifecycle | deciding rating behaviour or how the lifecycle should work |

## Relevant ADRs

| ADR | Decision | When |
|---|---|---|
| [ADR 0003](../../../docs/adr/0003-rating-lifecycle-determined-by-action.md) | the save action determines state, not prior state — save always lands `provisional`, including from `finalized` | questioning why saving a finalized album demotes it, or why finalize works from `unrated` |
