# Album Rating System — Design

**Date:** 2026-04-10
**Status:** Approved
**Replaces:** Existing 4-question rating questionnaire (`review/rating.go`)

---

## Overview

A personal album rating system that captures an evolving relationship with a record rather than a one-time verdict. Scores are anchored by 6 weighted base questions and refined by 3 modifiers, producing a 0.0–10.0 output to one decimal place.

The system has two active rating states (provisional and finalized) and one terminal incomplete state (stalled). Ratings are prompted on a 1-month cycle; snoozing delays re-prompt by 1 week per snooze, up to 3 snoozes before a record is marked stalled.

This fully replaces the existing 4-question curved questionnaire.

---

## Section 1: Scoring Engine

### Base Questions

6 questions, each normalized to a 1–5 scale. Default weight is equal (1/6 each), defined as named constants in code for easy per-question tuning.

| # | Key | Question | Scale | Default Weight |
|---|-----|----------|-------|----------------|
| 1 | `return_rate` | How often do you actively seek this record out? | 1–5 | `1/6` |
| 2 | `full_listen` | Do you listen front-to-back, or only return for select tracks? | Ternary → 1/3/5 | `1/6` |
| 3 | `emotional_resonance` | Does the music reliably move you across moods? | 1–5 | `1/6` |
| 4 | `sonic_pleasure` | Separate from meaning — do you simply like the way it sounds? | 1–5 | `1/6` |
| 5 | `recommendation_confidence` | Would you play this for someone whose taste you respect, without caveats? | Ternary → 1/3/5 | `1/6` |
| 6 | `shelf_test` | If you had to permanently delete it, would you actually care? | Binary → 1/5 | `1/6` |

Ternary questions present three options (low/mid/high) mapped to 1/3/5. Binary questions present two options (yes/no) mapped to 1/5.

### Score Formula

```
Base Score = weighted_average(answered questions) mapped linearly from [1,5] → [0,10]
Modifier Adjustment = average(modifier values) × max_swing_constant  (default: 0.75)
Final Score = clamp(Base Score + Modifier Adjustment, 0.0, 10.0), rounded to 1 decimal
```

**Provisional scoring:** Return Rate (Q1) is excluded; weighted average uses questions 2–6. Score is capped at 8.0 after the full formula runs.

**Max swing constant** is a named constant (`ModifierMaxSwing = 0.75`). Maximum possible swing: ±0.75. Mixed modifiers dampen each other naturally.

### Modifiers

3 modifiers, each ternary (−1/0/+1), unweighted and averaged together before applying the swing constant.

| Key | Modifier | +1 | 0 | −1 |
|-----|----------|----|---|----|
| `discovery_reward` | Discovery Reward | Deepens over time | Neutral | Felt exhausted quickly |
| `memorability` | Memorability | Highly sticky | Neutral | Doesn't stick at all |
| `life_association` | Life Association | Meaningful association | No strong history | Avoidant association |

### Confidence Flag

Checked after all answers are collected, before the confirm screen. Non-blocking — shows an interstitial with "Review answers" or "Proceed" options.

Contradiction patterns:
1. High Emotional Resonance (≥4) AND high Sonic Pleasure (≥4), but low Return Rate (≤2) — **finalized flow only** (Return Rate is not answered in provisional flow)
2. High Recommendation Confidence (≥4) but low Shelf Test (1)
3. High base score (≥7.0) but all modifiers negative (−1)

---

## Section 2: Database Schema

### New table: `album_rating_state`

Tracks the current lifecycle state per user+album. Created on first rating, always starting as `provisional`.

```sql
CREATE TABLE album_rating_state (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id       TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    state          TEXT NOT NULL CHECK(state IN ('provisional', 'finalized', 'stalled')),
    snooze_count   INTEGER NOT NULL DEFAULT 0,
    next_rerate_at DATETIME NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id)
);
```

### Modified table: `album_rating_log`

Add a `state` column to record which lifecycle state the rating was made in.

```sql
ALTER TABLE album_rating_log ADD COLUMN state TEXT CHECK(state IN ('provisional', 'finalized', 'stalled'));
```

Nullable for backwards compatibility — existing entries predate the state system and remain `NULL`. All new entries record state at time of rating.

Both changes ship as a single Goose migration.

---

## Section 3: Rerate Lifecycle

### State Transitions

| Event | From | To | Effect |
|-------|------|----|--------|
| First rating | — | `provisional` | Creates `album_rating_state`; `next_rerate_at` = now + 1 month |
| Snooze (snooze_count < 3) | `provisional` | `provisional` | `snooze_count++`; `next_rerate_at` += 1 week |
| Snooze (snooze_count = 3) | `provisional` | `stalled` | State set to `stalled`; no further rerate prompts |
| Update provisional | `provisional` | `provisional` | New log entry (state=provisional); `next_rerate_at` unchanged; snooze_count unchanged |
| Finalize | `provisional` or `stalled` | `finalized` | Full 6-question questionnaire; new log entry (state=finalized); no further rerate prompts |
| Re-rate finalized | `finalized` | `finalized` | Full 6-question questionnaire; new log entry (state=finalized); state record updated |

### Rerate Prompt Trigger

A rating is "due" when `next_rerate_at <= now()` and state is `provisional`. Stalled albums always appear in the carousel regardless of `next_rerate_at` — they are permanently due for finalization.

Clicking the rating of a due album opens the **rerate prompt screen** in the rating modal.

### Rerate Prompt Screen

Always two options, with the second option determined by whether the rating is due:

**When due** (`next_rerate_at <= now()`):
- **Rate now** → full 6-question flow → finalize
- **Snooze 1 week** → single HTMX POST, closes modal, updates snooze_count and next_rerate_at

**When not yet due** (provisional but `next_rerate_at` is in the future):
- **Rate now** → full 6-question flow → finalize
- **Update provisional** → 5-question flow (Return Rate excluded) → stays provisional, next_rerate_at unchanged

Clicking the rating of a provisional album always opens this prompt. Clicking the rating of a finalized album opens the confirm screen pre-filled with the current score (manual re-rating always possible).

**Stalled albums:** Clicking the rating (from the list or the carousel) opens a simplified rerate prompt with only **"Rate now"** (full 6-question flow → finalize). Snooze and Update Provisional are not offered — the record has exhausted its provisional period.

---

## Section 4: UI

### Questionnaire Flow (inside rating modal)

Multi-step, replacing the existing single-screen 4-question form:

1. **Base questions** — 5 radio-button fieldsets (provisional) or 6 (finalized). Same visual style as today.
2. **Modifiers** — second screen, 3 ternary radio groups (Positive / Neutral / Negative).
3. **Confidence check** (conditional) — interstitial if contradictions detected; "Review answers" returns to base questions, "Proceed" continues.
4. **Confirm screen** — pre-filled score (editable, 0–10 step 0.1), optional note, "Lock in" button. Same as today.

The **?** button on the confirm screen navigates back to base questions.

### Rating Color Coding

Applied to the rating number on the album list row and album detail page:

| State | Color | Tailwind class (approximate) |
|-------|-------|------------------------------|
| Finalized | Default | `text-base-content` |
| Provisional | Muted / washed out | `text-base-content/40` |
| Stalled | Muted reddish | `text-error/50` |
| Rerate due | Amber / call to action | `text-warning` |

Exact color values to be refined during implementation. The semantic mapping (finalized=default, provisional=muted, stalled=muted-red, due=amber) is fixed.

Stalled albums display a small "Stalled" badge alongside the rating.

### Dashboard Carousel

A new **"Rerate Due"** tab added to the existing carousel alongside "Recently Spun" and "Unrated". Shows: (1) provisional albums where `next_rerate_at <= now()`, and (2) all stalled albums. Each card: album art, title, artist, current score, state badge.

---

## Rating Scale

Unchanged from the existing system:

| Score | Label |
|-------|-------|
| 0–2.9 | DOA |
| 3.0–3.9 | Nope |
| 4.0–5.9 | Not For Me |
| 6.0–6.4 | Lukewarm |
| 6.5–6.9 | Solid |
| 7.0–7.9 | Staff Pick |
| 8.0–8.9 | Heavy Rotation |
| 9.0–9.9 | Instant Classic |
| 10.0 | Masterpiece |

---

## What Is Not Changing

- `album_rating_log` remains append-only; each new rating is a new row
- The confirm screen (score input + note) remains the final step in all flows
- The existing rating scale labels are unchanged
- Manual re-rating (bypassing the questionnaire) remains possible via the confirm screen
