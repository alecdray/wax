# Ratings

**What it is:** score albums 0–10 with a weighted questionnaire that measures your personal relationship with the record — not its critical merit.

## Why

Streaming platforms reduce your opinion to a binary like. Ratings give you a living score that evolves as your listening does — a number you can revise when you revisit an album years later. The questionnaire structure makes scoring deliberate, not impulsive.

## What a user can do

| Action | Result |
|---|---|
| Open the rating modal from any album surface | Score-entry form appears, pre-filled with your last rating if any. |
| Enter a score directly (0.0–10.0) | Save it provisionally or finalize it. |
| Use the questionnaire ("Help me score it") | Answer six weighted statements; a computed score fills the input. |
| Save a rating | The album enters `provisional` state. A provisional album is intended to be finalized later. |
| Save & finalize a rating | The album enters `finalized` state — your settled score. |
| Revise a finalized rating | Save again (back to provisional) or save & finalize with a new score. |

## Headline rules

- Scores are 0.0–10.0 to one decimal place.
- The questionnaire has six statements (three hard, three soft) answered on a Likert scale. Hard questions carry more weight — they're near-pass/fail signals. Soft questions measure degree of quality. Unanswered questions are excluded from the score.
- The target score is directionally accurate, not surgically precise. The questionnaire is guidance, not a formula.
- **Save** always lands the album in `provisional` — even if it was finalized before. **Save & finalize** always lands it in `finalized` — from any prior state.
- Every save or finalize writes a new rating-log entry. The history is preserved; the current state depends only on which action was last taken.
- Score labels describe a relationship (e.g. "it's fine," "this record shaped me"), not a letter grade.