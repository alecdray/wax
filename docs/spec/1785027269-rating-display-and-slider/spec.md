# Rating Display & Slider — Spec

## Goals

1. All saved rating values display as whole numbers (floor) everywhere in the UI.
2. The base-questions form replaces radio buttons with a continuous float slider per question.

## Non-goals

- No change to rating persistence (stored as float64).
- No change to the rating input form (`RatingConfirmFormFrag`) or its number input.
- No change to the confidence interstitial's displayed score.
- No change to the scoring formula.

---

## Change 1: Whole-number display

Add `FormatDisplayRating(r float64) string` to the `review` package:

```go
func FormatDisplayRating(r float64) string {
    return strconv.Itoa(int(math.Floor(r)))
}
```

Apply to all five display callsites (format strings → `review.FormatDisplayRating(...)`):

| File | Location | Current |
|---|---|---|
| `album_score_readout_frag.templ` | list-view large number | `"%.1f"` |
| `album_score_badge_frag.templ` | detail panel badge | `"%.1f"` |
| `album_rating_history_frag.templ` | history log score badge | `"%.4g"` |
| `carousel_section_frag.templ:79` | carousel card corner (regular) | `"%.1f"` |
| `carousel_section_frag.templ:127` | carousel card corner (provisional) | `"%.1f"` |

---

## Change 2: Question sliders

### Domain (`review.go`)

- `BaseQuestion.Value` changes from `int` to `float64`.
- `BaseQuestion.WithValue(v float64) BaseQuestion` updated to match.
- `Options []QuestionOption` removed from `BaseQuestion` (unused after slider migration).
- `BaseQuestions.Score()`: replace `float64(q.Value)` → `q.Value`; sentinel check `q.Value == 0` becomes `q.Value == 0.0` (no logic change, just type).
- `QuestionOption` type and `likertOptions` var removed.
- `AllBaseQuestions` entries drop the `Options` field.

### HTTP handler (`review/adapters/http.go`)

- Replace `strconv.Atoi(rawVal)` with `strconv.ParseFloat(rawVal, 64)`.
- Clamp parsed value to `[1.0, 5.0]` before calling `WithValue`.

### View (`review/adapters/views/base_questions_form_frag.templ`)

- Remove `baseQuestionRadio` and `baseQuestionFieldset`.
- New `baseQuestionSlider(q review.BaseQuestion)`:
  - DaisyUI `range` input: `class="range range-primary range-sm"`, `min="1"`, `max="5"`, `step="0.01"`, `value="3"` (neutral default), `name={ string(q.Key) }`, `required`.
  - `<datalist>` with five `<option>` elements at values 1–5 (no label text — browser renders ticks only); connected via `list` attribute.
  - `<legend>` question text unchanged above the slider.
- Unique datalist ID per question: `"ticks-" + string(q.Key)`.

### E2E tests (`e2e/spec/reviews.spec.ts`)

`answerQuestionnaire` updated to fill each `base-question-slider` range input with `'1'` instead of clicking `base-question-radio`.

### Additional fix (implementation)

`DetectContradictions` used `map[BaseQuestionKey]int` — updated to `map[BaseQuestionKey]float64` to match the `Value` type change.
