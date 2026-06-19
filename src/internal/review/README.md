# review

A personal album rating system that captures your ongoing relationship with a record — not a critical verdict, but a living score that evolves as your listening history does. Scores are 0.0–10.0 to one decimal place.

## Core philosophy

The rating system is personal, not critical. It measures your relationship with a record, not its objective quality or cultural significance. Scores are living values that evolve over time, not verdicts.

## Question design

- Questions are statements answered on a standard Likert scale (Strongly Disagree → Strongly Agree).
- Each statement should measure one thing only and be fast to answer — directional accuracy matters more than precision.
- Clarity beats personality, but personality is welcome where it doesn't sacrifice clarity.

## Hard vs soft questions

- **Hard questions** are near pass/fail. A low score on any of these is a strong signal the record doesn't belong in your collection. Higher weight.
- **Soft questions** measure degree of quality. A record can score modestly on these and still be worthy. Lower weight.

If a modifier feels important enough to weight, promote it to a base question instead.

## The base questions

The canonical question list lives in `AllBaseQuestions`. The rationale for each:

**Return Rate** *(Hard)* — Behavioural signal. Whether you'll keep coming back is a reliable proxy for how much you actually value a record.

**Track Quality** *(Soft)* — Consistency: how much of the record is good. Orthogonal to Sonic Pleasure: a record can be consistent in a genre you don't personally enjoy, or vice versa.

**Cohesion** *(Soft)* — Whether the record works as a complete piece. Split from Track Quality because they measure genuinely different things — a record can have great tracks but no cohesion, or strong cohesion with weaker tracks. Note that this question also captures composition: a record where every track sounds identical is technically cohesive but fails the "more than the sum of its parts" test — it's boring, and the question catches that.

**Emotional Resonance** *(Soft)* — Whether the record moves you. Soft because emotional response varies by mood and context more than the hard signals do.

**Sonic Pleasure** *(Hard)* — Purely personal taste, not quality. Intentionally captures subjective enjoyment rather than craft.

**Shelf Test** *(Hard)* — Alongside Return Rate, it requires lived experience to answer honestly. Saying no to a lot of records is the point: it reveals how much of your collection you're genuinely attached to versus keeping out of inertia.

## Scoring

The score is a weighted average of answered questions, linearly mapped to the 0–10 scale and rounded to one decimal place. Unanswered questions are excluded from both the weighted sum and the total weight, so partial ratings self-adjust without distorting the score.

## Rating lifecycle

The live state machine has two values: `provisional` and `finalized`. An album with no row in `album_rating_state` is `unrated` — there is no third stored value for it.

Application-driven transitions:

- `(no row) → provisional` — implicit, on the album's first saved rating.
- `provisional → finalized` — explicit, via the manual Finalize action on the score-entry form.
- `provisional → provisional` and `finalized → finalized` — re-rating an already-rated album appends a log entry and leaves the state value untouched.

There is no time-based promotion, no scheduled re-rate, and no snooze. Finalizing is the only path that promotes an album to `finalized`, and it only applies from `provisional`.

## Rating modal

The modal entry route (`GET /app/review/rating-recommender`) always returns the score-entry form (`RatingConfirmFormFrag`). The score input is pre-filled with the most-recent rating-log score when one exists, or left empty for an unrated album.

The questionnaire is opt-in — a button on the score-entry form opens it. Submitting the questionnaire computes a score from the answers and re-renders the score-entry form with that score pre-filled; dismissing it restores the prior pre-fill. The questionnaire never writes a rating row on its own.

The Finalize action is a second submit button on the score-entry form, visible only while the album is currently `provisional`. Clicking it saves the entered score and promotes the album to `finalized` in a single submission.

## Historical state values

The rating log (`album_rating_log`) preserves whatever lifecycle value was recorded at the time each entry was written. Entries written under earlier lifecycles can carry `stalled`; the column's CHECK constraint still admits it for that reason. `AlbumRatingHistoryFrag` renders log entries through `review.RatingStateLogLabel`, which has a label for the historical value alongside the live ones.

## Score labels

Each score maps to a label that describes a relationship with the record, not a letter grade — the spectrum runs from active rejection through faint praise to a top band reserved for genuinely transformative records.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
