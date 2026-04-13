# Review

A personal album rating system that captures your ongoing relationship with a record — not a critical verdict, but a living score that evolves as your listening history does. Scores are anchored by a set of base questions and refined by modifiers, producing a 0.0–10.0 output to one decimal place.

## Core Philosophy

The rating system is personal, not critical. It measures your relationship with a record, not its objective quality or cultural significance. Scores are living values that evolve over time, not verdicts.

## Question Design

- Questions should be fast and easy to answer — directional accuracy matters more than precision
- Each question should measure one thing only. If a question is secretly two questions, split it
- Questions and answers must be naturally aligned — the answers should feel like genuine responses to the question as worded
- Answers should feel like a natural internal monologue, not an analytical framework
- Clarity beats personality, but personality is welcome where it doesn't sacrifice clarity

**The scale:** All questions follow a none → some → most → all → transcendent arc. The transcendent tier (5) is always qualitatively different from the others, not just more of the same.

## Hard vs Soft Questions

- **Hard questions** are near pass/fail requirements. A 1 on any of these is a strong signal the record doesn't belong in your collection. Weight: 2.
- **Soft questions** measure degree of quality. A record can score modestly on these and still be worthy. Weight: 1.

If a modifier feels important enough to weight, it should be promoted to a base question instead.

## Base Questions

| Question | Type |
|---|---|
| Return Rate | Hard |
| Track Quality | Soft |
| Cohesion | Soft |
| Emotional Resonance | Hard |
| Sonic Pleasure | Hard |
| Shelf Test | Hard |

**Return Rate** *(Hard)* — Behavioural signal. How often you return is a reliable proxy for how much you actually value a record.

**Track Quality** *(Soft)* — How much of the record is good. Orthogonal to Sonic Pleasure: a record can have strong track quality in a genre you don't personally enjoy, or vice versa.

**Cohesion** *(Soft)* — Whether the record works as a complete piece. Split from Track Quality because they measure genuinely different things — a record can have great individual tracks but no cohesion, or strong cohesion with inconsistent track quality.

**Emotional Resonance** *(Hard)* — Uses confidence as a proxy for intensity, since dancing, crying, and contemplating life can't be put on the same scale. If you're certain you feel something, you score high. Strongly felt emotions and uncertainty are practically inversely correlated — if something hits hard, you know it. "Profoundly" at 5 collapses both certainty and intensity into one answer.

**Sonic Pleasure** *(Hard)* — Purely personal taste, not quality. A record can be well made but not resonate with you, or poorly made but speak to you completely. Intentionally captures subjective enjoyment rather than craft.

**Shelf Test** *(Hard)* — Ternary (No / I'd feel it / I'd be devastated). Finalized only — alongside Return Rate, it requires lived experience to answer honestly. Saying no to a lot of records is the point: it reveals how much of your collection you're genuinely attached to versus keeping out of inertia.

## Scoring

```
avg   = weightedSum / totalWeight
base  = (avg - 1.0) / 4.0 * 10.0
final = base + modifier_adjustment   [clamped 0–10, rounded to 1dp]
```

Unanswered questions are excluded from both sums, so partial ratings self-adjust without distorting the score.

## Modifiers

Modifiers are fast gut checks, not scored questions. They are ternary (−1 / 0 / +1), unweighted, and averaged together then multiplied by the max swing constant (default ±0.75). Their job is to nudge close calls, not reclassify records.

| Modifier | Purpose |
|---|---|
| Life Association | Is this record tied to a specific period or feeling in your life? |
| Interest | Do you find this record interesting beyond whether you enjoy it? |

## Rating Lifecycle

Ratings are provisional by default. Return Rate and Shelf Test are excluded from provisional scoring — both are attachment questions that require lived experience to answer honestly. The provisional set is purely about the music itself: Track Quality, Cohesion, Emotional Resonance, and Sonic Pleasure. Provisional scores are capped at 8.0.

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
| Provisional, rerate due | `base-content/40` | `badge-ghost opacity-40` | `badge-ghost text-base-content/60` |
| Finalized | `primary` | `badge-primary` | `badge-primary` |
| Stalled | `error/50` | `badge-error` | `badge-ghost text-base-content/60` |
