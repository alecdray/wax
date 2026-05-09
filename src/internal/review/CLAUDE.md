# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Single topic file `review.go` holds the `RatingState` enum, `RatingStateDTO` (per-album state machine — snoozing, rerate timing), `AlbumRatingDTO` (rating log entry), the scoring questionnaire, and the rating labels. `RatingState` is shared between the state machine and the log entry, so they live together.
