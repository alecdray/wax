# Review — domain

A personal album rating system measuring the user's relationship with a record, not its critical merit. Scores are living values that evolve over time.

## Rating values

- Scores are **0.0–10.0** to one decimal place.
- The score is derived from a weighted questionnaire of six statements (three hard, three soft), or entered directly.
- Score labels describe a relationship (e.g. "it's fine," "this record shaped me"), not a letter grade.

## Rating lifecycle state machine

The live state machine has two values: `provisional` and `finalized`. An album with no row in `album_rating_state` is `unrated` — there is no third stored value.

```
                 SaveRating          FinalizeWithRating
  +---------+  -------------->  +-------------+  -------------->  +-----------+
  | unrated |                   | provisional  |                   | finalized |
  +---------+  <--------------  +-------------+  <--------------  +-----------+
                 SaveRating          SaveRating
```

### Transitions

- **Save** (from any state): lands the album in `provisional`. Creates the state row on first rating; demotes a finalized album back to provisional. **Saving is the only path that un-finalizes.**
- **Finalize** (from any state): lands the album in `finalized`, whether the album was unrated, provisional, or already finalized.

Both actions append a new rating-log entry with the entered score. The resulting state depends **only on which action was taken**, never on where the album started.

## Rating log

The rating log (`album_rating_log`) is an **append-only history**. Every save or finalize writes a new entry with:
- The score entered
- The lifecycle state at the time of the entry
- An optional note

The log preserves whatever lifecycle value was recorded at write time. Entries written under earlier lifecycles can carry historical values (e.g. `stalled`); the column's CHECK constraint still admits those values because history is immutable.

## Questionnaire

The questionnaire computes a score from six weighted statements answered on a Likert scale (Strongly Disagree → Strongly Agree):

| Statement | Weight | Why it's weighted that way |
|---|---|---|
| Return Rate | Hard | Behavioural signal — whether you'll keep coming back. |
| Track Quality | Soft | Consistency: how much of the record is good. |
| Cohesion | Soft | Whether the record works as a complete piece. |
| Emotional Resonance | Soft | Whether the record moves you — varies by mood. |
| Sonic Pleasure | Hard | Purely personal taste, not quality. |
| Shelf Test | Hard | Requires lived experience — reveals genuine attachment vs. inertia. |

Unanswered questions are excluded from both the weighted sum and the total weight, so partial ratings self-adjust without distorting the score. The computed score is linearly mapped to 0–10 and rounded to one decimal place.

## Cross-domain effects

Rating saves and finalizations broadcast an `album-changed` HTMX event. The library domain owns the album surfaces that re-render in response — review's responsibility ends at writing the rating and signaling the change.