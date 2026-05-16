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

**Track Quality** *(Soft)* — How much of the record is good. Orthogonal to Sonic Pleasure: a record can have great tracks in a genre you don't personally enjoy, or vice versa.

**Cohesion** *(Soft)* — Whether the record works as a complete piece. Split from Track Quality because they measure genuinely different things — a record can have great tracks but no cohesion, or strong cohesion with weaker tracks.

**Emotional Resonance** *(Soft)* — Whether the record moves you. Soft because emotional response varies by mood and context more than the hard signals do.

**Sonic Pleasure** *(Hard)* — Purely personal taste, not quality. Intentionally captures subjective enjoyment rather than craft.

**Shelf Test** *(Hard)* — Alongside Return Rate, it requires lived experience to answer honestly. Saying no to a lot of records is the point: it reveals how much of your collection you're genuinely attached to versus keeping out of inertia.

## Scoring

The score is a weighted average of answered questions, linearly mapped to the 0–10 scale and rounded to one decimal place. Unanswered questions are excluded from both the weighted sum and the total weight, so partial ratings self-adjust without distorting the score.

## Rating lifecycle

Ratings are provisional by default — all six questions are presented regardless of mode. The mode is a marker on the rating itself, not a filter on the questionnaire.

After a cycle the record enters the rerate queue. You can finalize the rating, snooze it (which reschedules), or let it stall after the max snoozes is reached. Stalled records remain in the queue until acted on.

## Score labels

Each score maps to a label that describes a relationship with the record, not a letter grade — the spectrum runs from active rejection through faint praise to a top band reserved for genuinely transformative records.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
