# Review

A personal album rating system that captures your ongoing relationship with a record — not a critical verdict, but a living score that evolves as your listening history does. Scores are anchored by a set of base questions and refined by modifiers, producing a 0.0–10.0 output to one decimal place.

## Core Philosophy

The rating system is personal, not critical. It measures your relationship with a record, not its objective quality or cultural significance. Scores are living values that evolve over time, not verdicts.

## Question Design

- Questions are statements answered on a standard Likert scale: Strongly Disagree / Disagree / Neutral / Agree / Strongly Agree (1–5)
- Each statement should measure one thing only and be fast to answer — directional accuracy matters more than precision
- The Likert framing means statements just need to point at the right thing; the scale handles the gradation
- Clarity beats personality, but personality is welcome where it doesn't sacrifice clarity

## Hard vs Soft Questions

- **Hard questions** are near pass/fail requirements. A 1 on any of these is a strong signal the record doesn't belong in your collection. Weight: 2.
- **Soft questions** measure degree of quality. A record can score modestly on these and still be worthy. Weight: 1.

If a modifier feels important enough to weight, it should be promoted to a base question instead.

## Base Questions

| Statement | Type |
|---|---|
| I will keep coming back to this record | Hard |
| The tracks on this record consistently land | Soft |
| This record works as a complete piece | Soft |
| This record makes me feel something | Soft |
| I enjoy listening to this record | Hard |
| I would care if I had to permanently delete this record | Hard |

**Return Rate** *(Hard)* — Behavioural signal. Whether you'll keep coming back is a reliable proxy for how much you actually value a record. Forward-looking phrasing works for both provisional and finalized.

**Track Quality** *(Soft)* — How much of the record is good. Orthogonal to Sonic Pleasure: a record can have strong track quality in a genre you don't personally enjoy, or vice versa.

**Cohesion** *(Soft)* — Whether the record works as a complete piece. Split from Track Quality because they measure genuinely different things — a record can have great individual tracks but no cohesion, or strong cohesion with inconsistent track quality.

**Emotional Resonance** *(Soft)* — Whether the record moves you. Soft because emotional response varies by mood and context more than the hard signals do.

**Sonic Pleasure** *(Hard)* — Purely personal taste, not quality. A record can be well made but not resonate with you, or poorly made but speak to you completely. Intentionally captures subjective enjoyment rather than craft.

**Shelf Test** *(Hard)* — Alongside Return Rate, it requires lived experience to answer honestly. Saying no to a lot of records is the point: it reveals how much of your collection you're genuinely attached to versus keeping out of inertia.

## Scoring

```
avg   = weightedSum / totalWeight
base  = (avg - 1.0) / 4.0 * 10.0   [clamped 0–10, rounded to 1dp]
```

Unanswered questions are excluded from both sums, so partial ratings self-adjust without distorting the score.

## Rating Lifecycle

Ratings are provisional by default — all six questions are presented regardless of mode. The mode is a marker on the rating itself, not a filter on the questionnaire. Answering all questions for a provisional rating is encouraged; the score is the same calculation either way.

After a cycle (default 30 days), the record enters the rerate queue. You can finalize the rating, snooze it (reschedules for 7 days), or let it stall after three snoozes. Stalled records remain in the queue until acted on.

## Rating UI Elements

| Name | Component | Surface | Description |
|---|---|---|---|
| Score readout | `AlbumScoreReadout` | List view | Large prominent number, clickable to open rating modal |
| Score badge | `AlbumScoreBadge` | Detail panel | Compact `score - label` (e.g. `7.3 - Staff Pick`), clickable |
| Score label tag | — | List view metadata row | Non-interactive label-only display (`Staff Pick`, `Solid`, etc.) |

### Color coding

| State | Score readout | Score badge | Score label tag |
|---|---|---|---|
| No rating | `base-content/20` | `badge-primary` | `badge-ghost text-base-content/60` |
| Provisional | `base-content` | `badge-ghost` | `badge-ghost text-base-content/60` |
| Provisional, rerate due | `base-content/40` | `badge-ghost` | `badge-ghost text-base-content/60` |
| Finalized | `primary` | `badge-primary` | `badge-primary` |
| Stalled | `error/50` | `badge-error` | `badge-ghost text-base-content/60` |
