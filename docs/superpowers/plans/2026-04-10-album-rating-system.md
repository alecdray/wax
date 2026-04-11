# Album Rating System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing 4-question curved rating questionnaire with the spec'd 6-question + 3-modifier system, complete with provisional/finalized/stalled rating states, a monthly rerate lifecycle with snooze, and updated UI (color-coded ratings, rerate prompt modal, rerate carousel tab).

**Architecture:** New scoring engine lives entirely in `review/rating.go`; rating lifecycle state is tracked in a new `album_rating_state` DB table. The modal flow becomes multi-step (base questions → modifiers → optional confidence interstitial → confirm), driven by HTMX form chains with hidden inputs carrying computed values between steps.

**Tech Stack:** Go, SQLite + SQLC + Goose migrations, templ templates, HTMX, Tailwind CSS + DaisyUI. Use `task build/templ` after editing `.templ` files. Use `task db/up` after adding migrations (also regenerates schema.sql and runs sqlc). Use `task build/sqlc` after editing query files without a new migration.

---

## File Map

**New files:**
- `review/state.go` — `RatingState` string type, `RatingStateDTO` struct, `RatingMode` type
- `db/migrations/<timestamp>_album_rating_state.sql` — new table + state column on log
- `db/queries/album_rating_state.sql` — SQLC queries for `album_rating_state`

**Modified files:**
- `review/rating.go` — complete replacement: 6 questions, 3 modifiers, new formula, confidence flag
- `review/rating_test.go` — complete replacement for new engine
- `review/service.go` — update `AlbumRatingDTO`, update `AddRating` (accept state param), add `GetRatingState`, `UpsertRatingState`, `SnoozeRating`
- `db/queries/album_ratings.sql` — add `state` column to `InsertAlbumRatingLogEntry`
- `library/service.go` — add `RatingState *review.RatingStateDTO` to `AlbumDTO`, load states in `GetAlbumsInLibrary` and `GetAlbumInLibrary`, add `GetRerateQueue`
- `review/adapters/http.go` — state-aware `GetRatingRecommender`, add `SubmitModifiers`, `SnoozeRating` handlers; update `SubmitRatingRecommenderRating` to upsert state
- `review/adapters/rating.templ` — replace questionnaire templates; add modifier form, confidence interstitial, rerate prompt templates
- `library/adapters/dashboard.templ` — add `CarouselViewReratedue`, update `CarouselSection` (new tab), update `AlbumListRating` (state-based colors)
- `library/adapters/album_detail.templ` — update `AlbumRating` and `AlbumRatingHistory` (state colors, state badge in history)
- `server/server.go` — register `POST /app/review/rating-recommender/modifiers` and `POST /app/review/rating/snooze`

---

## Task 1: Replace Scoring Engine

**Files:**
- Rewrite: `src/internal/review/rating.go`
- Rewrite: `src/internal/review/rating_test.go`

- [ ] **Step 1: Write failing tests for the new engine**

Replace `src/internal/review/rating_test.go` entirely:

```go
package review

import (
	"math"
	"testing"
)

// --- BaseQuestions.Score ---

func TestBaseScore_AllMax_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-10.0) > 0.01 {
		t.Fatalf("expected 10.0 with all 5s (finalized), got %f", got)
	}
}

func TestBaseScore_AllMin_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 1
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-0.0) > 0.01 {
		t.Fatalf("expected 0.0 with all 1s (finalized), got %f", got)
	}
}

func TestBaseScore_AllMid_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-5.0) > 0.01 {
		t.Fatalf("expected 5.0 with all 3s (finalized), got %f", got)
	}
}

func TestBaseScore_Provisional_ExcludesReturnRate(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		if qs[i].Key == QuestionReturnRate {
			qs[i].Value = 1 // lowest possible
		} else {
			qs[i].Value = 5
		}
	}
	// provisional ignores return rate, so all remaining Qs are 5 → should be 10
	got := qs.Score(RatingModeProvisional)
	if math.Abs(got-10.0) > 0.01 {
		t.Fatalf("expected 10.0 (return rate excluded), got %f", got)
	}
}

func TestBaseScore_Provisional_CappedAt8(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	// All 5s in provisional mode: base=10 but capped at 8
	// (modifiers not included in Score, cap applied in FinalScore)
	score := qs.Score(RatingModeProvisional)
	if score > ProvisionalScoreCap {
		t.Fatalf("provisional base score %f exceeds cap %f", score, ProvisionalScoreCap)
	}
}

func TestBaseScore_IsRounded(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 2
	}
	got := qs.Score(RatingModeFinalized)
	rounded := math.Round(got*10) / 10
	if got != rounded {
		t.Fatalf("expected score rounded to 1dp, got %f", got)
	}
}

// --- Modifiers.Adjustment ---

func TestModifierAdjustment_AllPositive(t *testing.T) {
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = 1
	}
	got := mods.Adjustment()
	// average(1,1,1) * 0.75 = 0.75
	if math.Abs(got-ModifierMaxSwing) > 0.001 {
		t.Fatalf("expected +%f, got %f", ModifierMaxSwing, got)
	}
}

func TestModifierAdjustment_AllNegative(t *testing.T) {
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = -1
	}
	got := mods.Adjustment()
	if math.Abs(got-(-ModifierMaxSwing)) > 0.001 {
		t.Fatalf("expected -%f, got %f", ModifierMaxSwing, got)
	}
}

func TestModifierAdjustment_Mixed_Dampens(t *testing.T) {
	mods := defaultModifiers()
	mods[0].Value = 1
	mods[1].Value = -1
	mods[2].Value = 0
	got := mods.Adjustment()
	// average(1,-1,0) = 0 → 0 * 0.75 = 0
	if math.Abs(got) > 0.001 {
		t.Fatalf("expected 0 for mixed modifiers, got %f", got)
	}
}

// --- FinalScore ---

func TestFinalScore_ClampedAbove10(t *testing.T) {
	got := FinalScore(10.0, ModifierMaxSwing)
	if got > 10.0 {
		t.Fatalf("expected clamped to 10.0, got %f", got)
	}
}

func TestFinalScore_ClampedBelow0(t *testing.T) {
	got := FinalScore(0.0, -ModifierMaxSwing)
	if got < 0.0 {
		t.Fatalf("expected clamped to 0.0, got %f", got)
	}
}

// --- DetectContradictions ---

func TestDetectContradictions_HighERAndSP_LowRR_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionEmotionalResonance, QuestionSonicPleasure:
			qs[i].Value = 4
		case QuestionReturnRate:
			qs[i].Value = 2
		default:
			qs[i].Value = 3
		}
	}
	mods := defaultModifiers()
	if !DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected contradiction detected")
	}
}

func TestDetectContradictions_HighERAndSP_LowRR_Provisional_NoFlag(t *testing.T) {
	// contradiction 1 skipped in provisional (return rate not answered)
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionEmotionalResonance, QuestionSonicPleasure:
			qs[i].Value = 4
		default:
			qs[i].Value = 3
		}
	}
	mods := defaultModifiers()
	if DetectContradictions(qs, mods, 5.0, RatingModeProvisional) {
		t.Fatal("expected no contradiction in provisional (return rate excluded)")
	}
}

func TestDetectContradictions_HighRC_LowShelfTest(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionRecommendationConfidence:
			qs[i].Value = 5 // maps to value 5 (ternary)
		case QuestionShelfTest:
			qs[i].Value = 1 // binary: no
		default:
			qs[i].Value = 3
		}
	}
	mods := defaultModifiers()
	if !DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected contradiction detected")
	}
}

func TestDetectContradictions_HighScore_AllNegativeMods(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = -1
	}
	if !DetectContradictions(qs, mods, 7.5, RatingModeFinalized) {
		t.Fatal("expected contradiction: high base score + all negative mods")
	}
}

func TestDetectContradictions_NoContradiction(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	mods := defaultModifiers()
	if DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected no contradiction with mid scores and neutral mods")
	}
}

// --- GetRatingLabel ---

func TestGetRatingLabel_Ranges(t *testing.T) {
	cases := []struct {
		rating float64
		want   RatingLabel
	}{
		{0.0, RatingLabelDOA},
		{2.9, RatingLabelDOA},
		{3.0, RatingLabelNope},
		{3.9, RatingLabelNope},
		{4.0, RatingLabelNotForMe},
		{5.9, RatingLabelNotForMe},
		{6.0, RatingLabelLukewarm},
		{6.4, RatingLabelLukewarm},
		{6.5, RatingLabelSolid},
		{6.9, RatingLabelSolid},
		{7.0, RatingLabelRecommended},
		{7.9, RatingLabelRecommended},
		{8.0, RatingLabelEssential},
		{8.9, RatingLabelEssential},
		{9.0, RatingLabelInstantClassic},
		{9.9, RatingLabelInstantClassic},
		{10.0, RatingLabelMasterpiece},
	}
	for _, tc := range cases {
		got := GetRatingLabel(tc.rating)
		if got != tc.want {
			t.Errorf("GetRatingLabel(%v) = %q, want %q", tc.rating, got, tc.want)
		}
	}
}

// helpers

func finalizedQuestions() BaseQuestions {
	qs := make(BaseQuestions, len(AllBaseQuestions))
	copy(qs, AllBaseQuestions)
	return qs
}

func defaultModifiers() Modifiers {
	ms := make(Modifiers, len(AllModifiers))
	copy(ms, AllModifiers)
	return ms
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd /Users/shmoopy/workshop/projects/wax && go test ./src/internal/review/... -v 2>&1 | head -40
```

Expected: compilation errors (types don't exist yet).

- [ ] **Step 3: Write the new scoring engine**

Replace `src/internal/review/rating.go` entirely:

```go
package review

import (
	"math"

	"github.com/alecdray/wax/src/internal/core/utils"
)

// RatingMode controls which questions are active and whether the provisional cap applies.
type RatingMode string

const (
	RatingModeProvisional RatingMode = "provisional"
	RatingModeFinalized   RatingMode = "finalized"
)

// Score constants — adjust weights here to tune the system.
const (
	RatingWeightReturnRate               = 1.0 / 6.0
	RatingWeightFullListen               = 1.0 / 6.0
	RatingWeightEmotionalResonance       = 1.0 / 6.0
	RatingWeightSonicPleasure            = 1.0 / 6.0
	RatingWeightRecommendationConfidence = 1.0 / 6.0
	RatingWeightShelfTest                = 1.0 / 6.0

	// ModifierMaxSwing is the maximum total modifier adjustment (positive or negative).
	ModifierMaxSwing = 0.75

	// ProvisionalScoreCap is the maximum score a provisional rating can produce.
	ProvisionalScoreCap = 8.0
)

// BaseQuestionKey identifies a base question.
type BaseQuestionKey string

const (
	QuestionReturnRate               BaseQuestionKey = "return_rate"
	QuestionFullListen               BaseQuestionKey = "full_listen"
	QuestionEmotionalResonance       BaseQuestionKey = "emotional_resonance"
	QuestionSonicPleasure            BaseQuestionKey = "sonic_pleasure"
	QuestionRecommendationConfidence BaseQuestionKey = "recommendation_confidence"
	QuestionShelfTest                BaseQuestionKey = "shelf_test"
)

// QuestionOption is a single selectable answer for a base question.
type QuestionOption struct {
	Value int
	Label string
}

// BaseQuestion is a single base question in the rating questionnaire.
type BaseQuestion struct {
	Key      BaseQuestionKey
	Question string
	Options  []QuestionOption
	Weight   float64
	Value    int // 0 = unanswered
}

func (q BaseQuestion) WithValue(v int) BaseQuestion {
	q.Value = v
	return q
}

// BaseQuestions is a slice of BaseQuestion.
type BaseQuestions []BaseQuestion

// Score computes the base score (0–10) for the answered questions.
// In provisional mode, ReturnRate is excluded and the score is capped at ProvisionalScoreCap.
func (qs BaseQuestions) Score(mode RatingMode) float64 {
	var weightedSum, totalWeight float64
	for _, q := range qs {
		if mode == RatingModeProvisional && q.Key == QuestionReturnRate {
			continue
		}
		weightedSum += float64(q.Value) * q.Weight
		totalWeight += q.Weight
	}
	if totalWeight == 0 {
		return 0
	}
	avg := weightedSum / totalWeight
	// Map [1, 5] → [0, 10] linearly.
	base := (avg - 1.0) / 4.0 * 10.0
	if mode == RatingModeProvisional {
		base = math.Min(base, ProvisionalScoreCap)
	}
	return math.Round(base*10) / 10
}

// AllBaseQuestions is the canonical ordered list of base questions.
var AllBaseQuestions = BaseQuestions{
	{
		Key:      QuestionReturnRate,
		Question: "How often do you actively seek this record out?",
		Options: []QuestionOption{
			{1, "Almost never"},
			{2, "Rarely — only occasionally"},
			{3, "Sometimes — when the mood strikes"},
			{4, "Often — it comes to mind regularly"},
			{5, "Constantly — first thing I reach for"},
		},
		Weight: RatingWeightReturnRate,
	},
	{
		Key:      QuestionFullListen,
		Question: "Do you listen front-to-back, or only return for select tracks?",
		Options: []QuestionOption{
			{1, "Only a few tracks — skip most of it"},
			{3, "Mixed — some tracks, some full runs"},
			{5, "Front-to-back, every time"},
		},
		Weight: RatingWeightFullListen,
	},
	{
		Key:      QuestionEmotionalResonance,
		Question: "Does the music reliably move you? Does it hit differently across moods?",
		Options: []QuestionOption{
			{1, "Not at all — leaves me cold"},
			{2, "Occasionally, in the right mood"},
			{3, "Yes, in a consistent but modest way"},
			{4, "Reliably moves me across different moods"},
			{5, "Deeply — hits differently every time"},
		},
		Weight: RatingWeightEmotionalResonance,
	},
	{
		Key:      QuestionSonicPleasure,
		Question: "Separate from meaning — do you simply like the way it sounds?",
		Options: []QuestionOption{
			{1, "No — the sound is actively off-putting"},
			{2, "Not really — it doesn't sound appealing"},
			{3, "It's fine — nothing special sonically"},
			{4, "Yes — I like the way it sounds"},
			{5, "Absolutely — it sounds incredible"},
		},
		Weight: RatingWeightSonicPleasure,
	},
	{
		Key:      QuestionRecommendationConfidence,
		Question: "Would you play this for someone whose taste you respect, without caveats?",
		Options: []QuestionOption{
			{1, "No — I'd be embarrassed"},
			{3, "Maybe — with some caveats"},
			{5, "Yes — without hesitation"},
		},
		Weight: RatingWeightRecommendationConfidence,
	},
	{
		Key:      QuestionShelfTest,
		Question: "If you had to permanently delete it, would you actually care?",
		Options: []QuestionOption{
			{1, "No — wouldn't miss it"},
			{5, "Yes — I'd genuinely care"},
		},
		Weight: RatingWeightShelfTest,
	},
}

// ModifierKey identifies a modifier.
type ModifierKey string

const (
	ModifierDiscoveryReward ModifierKey = "discovery_reward"
	ModifierMemorability    ModifierKey = "memorability"
	ModifierLifeAssociation ModifierKey = "life_association"
)

// ModifierOption is a selectable value for a modifier.
type ModifierOption struct {
	Value int // -1, 0, or +1
	Label string
}

// Modifier is a single gut-check adjustment applied on top of the base score.
type Modifier struct {
	Key     ModifierKey
	Label   string
	Options []ModifierOption
	Value   int // -1, 0, or +1
}

func (m Modifier) WithValue(v int) Modifier {
	m.Value = v
	return m
}

// Modifiers is a slice of Modifier.
type Modifiers []Modifier

// Adjustment computes the total modifier adjustment: average(values) × ModifierMaxSwing.
func (ms Modifiers) Adjustment() float64 {
	if len(ms) == 0 {
		return 0
	}
	var sum float64
	for _, m := range ms {
		sum += float64(m.Value)
	}
	return (sum / float64(len(ms))) * ModifierMaxSwing
}

// AllModifiers is the canonical list of modifiers.
var AllModifiers = Modifiers{
	{
		Key:   ModifierDiscoveryReward,
		Label: "Discovery Reward",
		Options: []ModifierOption{
			{1, "Deepens over time"},
			{0, "Neutral"},
			{-1, "Felt exhausted quickly"},
		},
	},
	{
		Key:   ModifierMemorability,
		Label: "Memorability",
		Options: []ModifierOption{
			{1, "Highly sticky — stays with me"},
			{0, "Neutral"},
			{-1, "Doesn't stick at all"},
		},
	},
	{
		Key:   ModifierLifeAssociation,
		Label: "Life Association",
		Options: []ModifierOption{
			{1, "Meaningful association"},
			{0, "No strong history"},
			{-1, "Avoidant association"},
		},
	},
}

// FinalScore clamps and rounds the combined base score + modifier adjustment.
func FinalScore(baseScore, modifierAdjustment float64) float64 {
	return math.Round(utils.Clamp(baseScore+modifierAdjustment, 0.0, 10.0)*10) / 10
}

// DetectContradictions returns true if the answers contain internally contradictory signals.
// Contradiction 1 (finalized only): high Emotional Resonance AND Sonic Pleasure, but low Return Rate.
// Contradiction 2: high Recommendation Confidence but low Shelf Test.
// Contradiction 3: high base score but all modifiers negative.
func DetectContradictions(qs BaseQuestions, mods Modifiers, baseScore float64, mode RatingMode) bool {
	qByKey := make(map[BaseQuestionKey]int, len(qs))
	for _, q := range qs {
		qByKey[q.Key] = q.Value
	}

	// Contradiction 1 — finalized only
	if mode == RatingModeFinalized {
		er := qByKey[QuestionEmotionalResonance]
		sp := qByKey[QuestionSonicPleasure]
		rr := qByKey[QuestionReturnRate]
		if er >= 4 && sp >= 4 && rr <= 2 {
			return true
		}
	}

	// Contradiction 2
	rc := qByKey[QuestionRecommendationConfidence]
	st := qByKey[QuestionShelfTest]
	if rc >= 4 && st == 1 {
		return true
	}

	// Contradiction 3
	if baseScore >= 7.0 {
		allNeg := true
		for _, m := range mods {
			if m.Value != -1 {
				allNeg = false
				break
			}
		}
		if allNeg {
			return true
		}
	}

	return false
}

// --- Rating labels (unchanged) ---

type RatingLabel string

const (
	RatingLabelDOA            RatingLabel = "DOA"
	RatingLabelNope           RatingLabel = "Nope"
	RatingLabelNotForMe       RatingLabel = "Not For Me"
	RatingLabelLukewarm       RatingLabel = "Lukewarm"
	RatingLabelSolid          RatingLabel = "Solid"
	RatingLabelRecommended    RatingLabel = "Staff Pick"
	RatingLabelEssential      RatingLabel = "Heavy Rotation"
	RatingLabelInstantClassic RatingLabel = "Instant Classic"
	RatingLabelMasterpiece    RatingLabel = "Masterpiece"
)

type RatingKeyEntry struct {
	MinValue float64
	MaxValue float64
	Label    RatingLabel
}

var RatingKey = []RatingKeyEntry{
	{0.0, 2.9, RatingLabelDOA},
	{3.0, 3.9, RatingLabelNope},
	{4.0, 5.9, RatingLabelNotForMe},
	{6.0, 6.4, RatingLabelLukewarm},
	{6.5, 6.9, RatingLabelSolid},
	{7.0, 7.9, RatingLabelRecommended},
	{8.0, 8.9, RatingLabelEssential},
	{9.0, 9.9, RatingLabelInstantClassic},
	{10.0, 10.0, RatingLabelMasterpiece},
}

func GetRatingLabel(rating float64) RatingLabel {
	clamped := utils.Clamp(rating, 0, 10)
	for _, entry := range RatingKey {
		if clamped >= entry.MinValue && clamped <= entry.MaxValue {
			return entry.Label
		}
	}
	return RatingKey[len(RatingKey)-1].Label
}
```

- [ ] **Step 4: Run tests — verify they pass**

```bash
cd /Users/shmoopy/workshop/projects/wax && go test ./src/internal/review/... -v 2>&1
```

Expected: all tests PASS. Fix any failures before continuing.

- [ ] **Step 5: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/review/rating.go src/internal/review/rating_test.go && git commit -m "feat(review): replace 4-question engine with 6Q+3M scoring system"
```

---

## Task 2: Rating State Type

**Files:**
- Create: `src/internal/review/state.go`

- [ ] **Step 1: Create state.go**

```go
package review

import "time"

// RatingState represents the lifecycle state of an album rating.
type RatingState string

const (
	RatingStateProvisional RatingState = "provisional"
	RatingStateFinalized   RatingState = "finalized"
	RatingStateStalled     RatingState = "stalled"
)

// RatingStateDTO carries the current lifecycle state for a user+album pair.
type RatingStateDTO struct {
	ID           string
	UserID       string
	AlbumID      string
	State        RatingState
	SnoozeCount  int
	NextRerateAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsReratedue returns true when the album should be prompted for a rerate.
// Provisional: due when next_rerate_at is in the past.
// Stalled: always due.
func (s *RatingStateDTO) IsRerateDue() bool {
	if s.State == RatingStateStalled {
		return true
	}
	return s.State == RatingStateProvisional && !s.NextRerateAt.After(time.Now())
}
```

- [ ] **Step 2: Confirm it compiles**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/review/state.go && git commit -m "feat(review): add RatingState and RatingStateDTO types"
```

---

## Task 3: Database Migration

**Files:**
- Create: `db/migrations/<timestamp>_album_rating_state.sql` (use `task db/create -- album_rating_state`)

- [ ] **Step 1: Create migration file**

```bash
cd /Users/shmoopy/workshop/projects/wax && task db/create -- album_rating_state
```

This creates a file in `db/migrations/` with a timestamp prefix. Open that file and replace its contents with:

```sql
-- +goose Up
-- +goose StatementBegin
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

ALTER TABLE album_rating_log ADD COLUMN state TEXT CHECK(state IN ('provisional', 'finalized', 'stalled'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE album_rating_state;
-- note: SQLite does not support DROP COLUMN; state column on album_rating_log intentionally left
-- +goose StatementEnd
```

- [ ] **Step 2: Run migration**

```bash
cd /Users/shmoopy/workshop/projects/wax && task db/up
```

Expected: migration applies, `db/schema.sql` updated, sqlc regenerates without error.

- [ ] **Step 3: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add db/migrations/ db/schema.sql src/internal/core/db/sqlc/ && git commit -m "feat(db): add album_rating_state table and state column to rating log"
```

---

## Task 4: SQLC Queries for Rating State

**Files:**
- Create: `db/queries/album_rating_state.sql`
- Modify: `db/queries/album_ratings.sql`

- [ ] **Step 1: Create album_rating_state.sql**

```sql
-- name: InsertAlbumRatingState :one
INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at, created_at, updated_at)
VALUES (?, ?, ?, ?, 0, ?, current_timestamp, current_timestamp)
RETURNING *;

-- name: UpdateAlbumRatingState :one
UPDATE album_rating_state
SET state = ?, snooze_count = ?, next_rerate_at = ?, updated_at = current_timestamp
WHERE user_id = ? AND album_id = ?
RETURNING *;

-- name: GetAlbumRatingState :one
SELECT * FROM album_rating_state
WHERE user_id = ? AND album_id = ?;

-- name: GetAllAlbumRatingStates :many
SELECT * FROM album_rating_state
WHERE user_id = ?;

-- name: GetRerateQueueAlbums :many
SELECT
    albums.id,
    albums.spotify_id,
    albums.title,
    albums.image_url,
    COALESCE((
        SELECT GROUP_CONCAT(ar.name, ', ')
        FROM (SELECT DISTINCT ar2.id, ar2.name FROM album_artists aa JOIN artists ar2 ON ar2.id = aa.artist_id WHERE aa.album_id = albums.id) AS ar
    ), '') AS artist_names,
    ars.state,
    arl.rating
FROM album_rating_state ars
JOIN albums ON albums.id = ars.album_id
LEFT JOIN (
    SELECT arl2.album_id, arl2.rating
    FROM album_rating_log arl2
    JOIN (
        SELECT album_id, MAX(created_at) AS max_created_at
        FROM album_rating_log
        WHERE user_id = ?
        GROUP BY album_id
    ) latest ON arl2.album_id = latest.album_id AND arl2.created_at = latest.max_created_at
    WHERE arl2.user_id = ?
) arl ON arl.album_id = ars.album_id
WHERE ars.user_id = ?
  AND (
    (ars.state = 'provisional' AND ars.next_rerate_at <= current_timestamp)
    OR ars.state = 'stalled'
  )
ORDER BY ars.state DESC, ars.next_rerate_at ASC;
```

- [ ] **Step 2: Update album_ratings.sql — add state to InsertAlbumRatingLogEntry**

In `db/queries/album_ratings.sql`, replace the `InsertAlbumRatingLogEntry` query:

```sql
-- name: InsertAlbumRatingLogEntry :one
INSERT INTO album_rating_log (id, user_id, album_id, rating, note, state, created_at)
VALUES (?, ?, ?, ?, ?, ?, current_timestamp)
RETURNING *;
```

- [ ] **Step 3: Regenerate SQLC**

```bash
cd /Users/shmoopy/workshop/projects/wax && task build/sqlc
```

Expected: generates updated files in `src/internal/core/db/sqlc/` with no errors.

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/... 2>&1
```

Expected: compilation errors in `review/service.go` because `InsertAlbumRatingLogEntryParams` now requires a `State` field. Note the errors — they will be fixed in the next task.

- [ ] **Step 5: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add db/queries/ src/internal/core/db/sqlc/ && git commit -m "feat(db): add SQLC queries for album_rating_state and update rating log insert"
```

---

## Task 5: Update Review Service

**Files:**
- Modify: `src/internal/review/service.go`

- [ ] **Step 1: Replace service.go**

Replace the entire file:

```go
package review

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/google/uuid"
)

type AlbumRatingDTO struct {
	ID        string
	UserID    string
	AlbumID   string
	Rating    *float64
	Note      *string
	State     *RatingState
	CreatedAt time.Time
}

func NewAlbumRatingDTOFromModel(model sqlc.AlbumRatingLog) *AlbumRatingDTO {
	dto := &AlbumRatingDTO{
		ID:        model.ID,
		UserID:    model.UserID,
		AlbumID:   model.AlbumID,
		CreatedAt: model.CreatedAt,
		Rating:    &model.Rating,
	}
	if model.Note.Valid {
		dto.Note = &model.Note.String
	}
	if model.State.Valid {
		s := RatingState(model.State.String)
		dto.State = &s
	}
	return dto
}

func NewRatingStateDTOFromModel(model sqlc.AlbumRatingState) *RatingStateDTO {
	return &RatingStateDTO{
		ID:           model.ID,
		UserID:       model.UserID,
		AlbumID:      model.AlbumID,
		State:        RatingState(model.State),
		SnoozeCount:  int(model.SnoozeCount),
		NextRerateAt: model.NextRerateAt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// AddRating appends a new rating log entry with the given state.
func (s *Service) AddRating(ctx context.Context, userID, albumID string, rating float64, note string, state RatingState) (*AlbumRatingDTO, error) {
	var noteParam sql.NullString
	if note != "" {
		noteParam = sql.NullString{String: note, Valid: true}
	}
	model, err := s.db.Queries().InsertAlbumRatingLogEntry(ctx, sqlc.InsertAlbumRatingLogEntryParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		AlbumID: albumID,
		Rating:  rating,
		Note:    noteParam,
		State:   sql.NullString{String: string(state), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return NewAlbumRatingDTOFromModel(model), nil
}

// DeleteRatingEntry removes a single rating log entry.
func (s *Service) DeleteRatingEntry(ctx context.Context, userID, entryID string) error {
	return s.db.Queries().DeleteAlbumRatingLogEntry(ctx, sqlc.DeleteAlbumRatingLogEntryParams{
		ID:     entryID,
		UserID: userID,
	})
}

// GetRatingLog returns all rating log entries for a user+album, newest first.
func (s *Service) GetRatingLog(ctx context.Context, userID, albumID string) ([]*AlbumRatingDTO, error) {
	rows, err := s.db.Queries().GetUserAlbumRatingLog(ctx, sqlc.GetUserAlbumRatingLogParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	dtos := make([]*AlbumRatingDTO, len(rows))
	for i, row := range rows {
		dtos[i] = NewAlbumRatingDTOFromModel(row)
	}
	return dtos, nil
}

// GetRatingState returns the current lifecycle state for a user+album, or nil if never rated.
func (s *Service) GetRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	model, err := s.db.Queries().GetAlbumRatingState(ctx, sqlc.GetAlbumRatingStateParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

// GetAllRatingStates returns all rating states for a user, keyed by album ID.
func (s *Service) GetAllRatingStates(ctx context.Context, userID string) (map[string]*RatingStateDTO, error) {
	rows, err := s.db.Queries().GetAllAlbumRatingStates(ctx, userID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*RatingStateDTO, len(rows))
	for _, row := range rows {
		dto := NewRatingStateDTOFromModel(row)
		m[row.AlbumID] = dto
	}
	return m, nil
}

// CreateRatingState creates the initial provisional state for a newly rated album.
func (s *Service) CreateRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	model, err := s.db.Queries().InsertAlbumRatingState(ctx, sqlc.InsertAlbumRatingStateParams{
		ID:           uuid.NewString(),
		UserID:       userID,
		AlbumID:      albumID,
		State:        string(RatingStateProvisional),
		NextRerateAt: time.Now().AddDate(0, 1, 0),
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

// FinalizeRating transitions the album rating state to finalized.
func (s *Service) FinalizeRating(ctx context.Context, userID, albumID string, current *RatingStateDTO) (*RatingStateDTO, error) {
	snooze := 0
	if current != nil {
		snooze = current.SnoozeCount
	}
	model, err := s.db.Queries().UpdateAlbumRatingState(ctx, sqlc.UpdateAlbumRatingStateParams{
		State:        string(RatingStateFinalized),
		SnoozeCount:  int64(snooze),
		NextRerateAt: time.Now(),
		UserID:       userID,
		AlbumID:      albumID,
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

// SnoozeRating increments snooze count and bumps next_rerate_at by 1 week.
// After 3 snoozes the state transitions to stalled.
const maxSnoozeCount = 3

func (s *Service) SnoozeRating(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	current, err := s.GetRatingState(ctx, userID, albumID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("no rating state found for album")
	}

	newSnooze := current.SnoozeCount + 1
	newState := RatingStateProvisional
	newNextRerate := current.NextRerateAt.AddDate(0, 0, 7)

	if newSnooze >= maxSnoozeCount {
		newState = RatingStateStalled
	}

	model, err := s.db.Queries().UpdateAlbumRatingState(ctx, sqlc.UpdateAlbumRatingStateParams{
		State:        string(newState),
		SnoozeCount:  int64(newSnooze),
		NextRerateAt: newNextRerate,
		UserID:       userID,
		AlbumID:      albumID,
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/... 2>&1
```

Expected: compilation errors in `review/adapters/http.go` because `AddRating` signature changed. Note the errors — fixed in Task 7.

- [ ] **Step 3: Run review tests**

```bash
cd /Users/shmoopy/workshop/projects/wax && go test ./src/internal/review/... -v 2>&1
```

Expected: all tests PASS (tests don't use `service.go`).

- [ ] **Step 4: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/review/service.go src/internal/review/state.go && git commit -m "feat(review): update service with state management and new AddRating signature"
```

---

## Task 6: Add RatingState to Library Service

**Files:**
- Modify: `src/internal/library/service.go`

- [ ] **Step 1: Add RatingState field to AlbumDTO**

In `src/internal/library/service.go`, find the `AlbumDTO` struct (around line 110) and add the `RatingState` field:

```go
type AlbumDTO struct {
	ID           string
	SpotifyID    string
	Title        string
	ImageURL     string
	Artists      []ArtistDTO
	Tracks       []TrackDTO
	Releases     ReleaseDTOs
	Rating       *review.AlbumRatingDTO
	RatingLog    []*review.AlbumRatingDTO
	RatingState  *review.RatingStateDTO
	Tags         []tags.TagDTO
	SleeveNote   *notes.AlbumNoteDTO
	LastPlayedAt *time.Time
}
```

- [ ] **Step 2: Load rating states in GetAlbumsInLibrary**

In `GetAlbumsInLibrary`, after the `ratings` fetch block (after line ~456), add a rating states fetch:

```go
ratingStates, err := s.reviewService.GetAllRatingStates(ctx, userId)
if err != nil {
    return nil, fmt.Errorf("failed to get rating states: %w", err)
}
```

Then inside the album DTO construction loop, after `dto.Tags = ...`, add:

```go
dto.RatingState = ratingStates[album.ID]
```

- [ ] **Step 3: Load rating state in GetAlbumInLibrary**

In `GetAlbumInLibrary`, after the `ratingLog` fetch block (around line 670), add:

```go
ratingState, err := s.reviewService.GetRatingState(ctx, userId, albumId)
if err != nil {
    return nil, fmt.Errorf("failed to get rating state: %w", err)
}
```

Then after `albumDto.RatingLog = ratingLog`, add:

```go
albumDto.RatingState = ratingState
```

- [ ] **Step 4: Add GetRerateQueue to library service**

Add this method to `service.go`. It returns albums due for rerate, including stalled ones.

First, add a new `RerateAlbumDTO` type near the other DTO types:

```go
type RerateAlbumDTO struct {
	ID          string
	SpotifyID   string
	Title       string
	Artists     string
	ImageURL    string
	Rating      *float64
	RatingState review.RatingState
}
```

Then add the method. Check what fields the `GetRerateQueueAlbums` SQLC query returns — it returns `id, spotify_id, title, image_url, artist_names, state, rating`. The exact Go struct name will be `GetRerateQueueAlbumsRow` in the generated sqlc code. Verify with:

```bash
grep -A 10 "GetRerateQueueAlbumsRow" /Users/shmoopy/workshop/projects/wax/src/internal/core/db/sqlc/album_rating_state.sql.go
```

Then add the method:

```go
func (s *Service) GetRerateQueue(ctx context.Context, userID string) ([]RerateAlbumDTO, error) {
	rows, err := s.db.Queries().GetRerateQueueAlbums(ctx, sqlc.GetRerateQueueAlbumsParams{
		UserID:   userID,
		UserID_2: userID,
		UserID_3: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rerate queue: %w", err)
	}
	dtos := make([]RerateAlbumDTO, 0, len(rows))
	for _, row := range rows {
		dto := RerateAlbumDTO{
			ID:          row.ID,
			SpotifyID:   row.SpotifyID,
			Title:       row.Title,
			Artists:     fmt.Sprintf("%s", row.ArtistNames),
			ImageURL:    row.ImageUrl.String,
			RatingState: review.RatingState(row.State),
		}
		if row.Rating.Valid {
			dto.Rating = &row.Rating.Float64
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}
```

> **Note:** The SQLC-generated params struct name for `GetRerateQueueAlbums` may differ depending on how many `?` params it uses. If it uses three `user_id` params, the struct will have `UserID`, `UserID_2`, `UserID_3` fields. Verify with `grep -A 5 "GetRerateQueueAlbumsParams" src/internal/core/db/sqlc/album_rating_state.sql.go`. Adjust field names to match.

- [ ] **Step 5: Wire reviewService into library.Service**

Check if `library.Service` already holds a `reviewService` field:

```bash
grep -n "reviewService\|review.Service\|review.NewService" /Users/shmoopy/workshop/projects/wax/src/internal/library/service.go | head -10
```

If it does not, find the `Service` struct definition and add it, then update `NewService` to accept and store it. Also update the call site in `server/server.go`. If it already exists, skip this step.

- [ ] **Step 6: Verify compilation**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/... 2>&1
```

Fix any remaining errors.

- [ ] **Step 7: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/library/service.go && git commit -m "feat(library): add RatingState to AlbumDTO, load states, add GetRerateQueue"
```

---

## Task 7: Rerate Carousel Tab

**Files:**
- Modify: `src/internal/library/adapters/dashboard.templ`
- Modify: `src/internal/library/adapters/http.go`

- [ ] **Step 1: Add CarouselViewReratedue and rerate carousel strip to dashboard.templ**

Add the new carousel view constant alongside the existing ones:

```go
const (
	CarouselViewRecentlyPlayed CarouselView = "recently-played"
	CarouselViewUnrated        CarouselView = "unrated"
	CarouselViewReratedue      CarouselView = "rerate-due"
)
```

Add a new carousel strip template for rerate items (after `carouselStrip`):

```go
templ rerateCarouselStrip(albums []library.RerateAlbumDTO) {
	if len(albums) == 0 {
		<div class="px-4 py-4 text-xs text-base-content/40">No albums due for rerate</div>
	} else {
		<div class="carousel carousel-end gap-3 px-4 py-2 w-full overscroll-x-none">
			for _, album := range albums {
				<div class="carousel-item">
					<a
						href={ templ.URL(fmt.Sprintf("/app/library/albums/%s", album.ID)) }
						class="flex flex-col items-center gap-1 hover:opacity-80 transition-opacity w-26"
					>
						if album.ImageURL != "" {
							<div class="avatar">
								<div class="mask mask-squircle h-24 w-24 flex-shrink-0">
									<img src={ album.ImageURL } alt={ album.Title }/>
								</div>
							</div>
						}
						<span class="text-xs text-nowrap truncate w-full text-left">{ album.Title }</span>
						if album.Artists != "" {
							<span class="text-xs text-nowrap truncate w-full text-left text-base-content/40">{ album.Artists }</span>
						}
						<div class="flex items-center gap-1 w-full">
							if album.Rating != nil {
								<span class="text-xs font-bold tabular-nums">{ fmt.Sprintf("%.1f", *album.Rating) }</span>
							}
							if album.RatingState == "stalled" {
								<span class="badge badge-xs badge-error badge-soft">Stalled</span>
							}
						</div>
					</a>
				</div>
			}
		</div>
	}
}
```

Update `CarouselSection` to accept rerate albums and add the new tab. Replace the current `CarouselSection` signature and body:

```go
type CarouselSectionProps struct {
	RegularAlbums []library.AlbumSummaryDTO
	RerateAlbums  []library.RerateAlbumDTO
	Active        CarouselView
}

templ CarouselSection(props CarouselSectionProps) {
	<div
		id="carousel-section"
		class="w-full flex-shrink-0"
		hx-get="/app/library/dashboard/carousel"
		hx-trigger="libraryUpdated from:body"
		hx-swap="outerHTML"
	>
		<div class="flex items-center gap-3 px-4 pb-1">
			<button
				class={ "text-xs font-semibold uppercase tracking-widest transition-colors", templ.KV("text-base-content", props.Active == CarouselViewRecentlyPlayed), templ.KV("text-base-content/40 hover:text-base-content/70 cursor-pointer", props.Active != CarouselViewRecentlyPlayed) }
				if props.Active != CarouselViewRecentlyPlayed {
					hx-get="/app/library/dashboard/carousel?view=recently-played"
					hx-target="#carousel-section"
					hx-swap="outerHTML"
				}
				data-testid="carousel-recently-spun-tab"
			>Recently Spun</button>
			<span class="text-base-content/20 cursor-default">|</span>
			<button
				class={ "text-xs font-semibold uppercase tracking-widest transition-colors", templ.KV("text-base-content", props.Active == CarouselViewUnrated), templ.KV("text-base-content/40 hover:text-base-content/70 cursor-pointer", props.Active != CarouselViewUnrated) }
				if props.Active != CarouselViewUnrated {
					hx-get="/app/library/dashboard/carousel?view=unrated"
					hx-target="#carousel-section"
					hx-swap="outerHTML"
				}
				data-testid="carousel-unrated-tab"
			>Unrated</button>
			<span class="text-base-content/20 cursor-default">|</span>
			<button
				class={ "text-xs font-semibold uppercase tracking-widest transition-colors", templ.KV("text-base-content", props.Active == CarouselViewReratedue), templ.KV("text-base-content/40 hover:text-base-content/70 cursor-pointer", props.Active != CarouselViewReratedue) }
				if props.Active != CarouselViewReratedue {
					hx-get="/app/library/dashboard/carousel?view=rerate-due"
					hx-target="#carousel-section"
					hx-swap="outerHTML"
				}
				data-testid="carousel-rerate-due-tab"
			>Rerate Due</button>
		</div>
		switch props.Active {
		case CarouselViewUnrated:
			@carouselStrip(props.RegularAlbums, "No unrated albums in your library")
		case CarouselViewReratedue:
			@rerateCarouselStrip(props.RerateAlbums)
		default:
			@carouselStrip(props.RegularAlbums, "No recently played albums")
		}
	</div>
}
```

Update all call sites of `CarouselSection` in `dashboard.templ`. Find the existing call in `DashboardPage` and update it to pass a `CarouselSectionProps{}`. The initial load uses recently-played albums; rerate albums are nil on initial load (loaded on tab click):

```go
@CarouselSection(CarouselSectionProps{
    RegularAlbums: props.RecentAlbums,
    Active:        CarouselViewRecentlyPlayed,
})
```

- [ ] **Step 2: Update GetCarousel handler in library/adapters/http.go**

Find `GetCarousel` and update it to handle the new view. The `h.libraryService` needs a `GetRerateQueue` method (added in Task 6).

```go
func (h *HttpHandler) GetCarousel(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := CarouselView(r.URL.Query().Get("view"))
	if view == "" {
		view = CarouselViewRecentlyPlayed
	}

	props := CarouselSectionProps{Active: view}

	switch view {
	case CarouselViewUnrated:
		albums, err := h.libraryService.GetUnratedAlbums(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.RegularAlbums = albums
	case CarouselViewReratedue:
		albums, err := h.libraryService.GetRerateQueue(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.RerateAlbums = albums
	default:
		props.Active = CarouselViewRecentlyPlayed
		albums, err := h.libraryService.GetRecentlyPlayedAlbums(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.RegularAlbums = albums
	}

	CarouselSection(props).Render(r.Context(), w)
}
```

- [ ] **Step 3: Build templates**

```bash
cd /Users/shmoopy/workshop/projects/wax && task build/templ 2>&1
```

Expected: no errors.

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/... 2>&1
```

- [ ] **Step 5: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/library/adapters/ && git commit -m "feat(library): add Rerate Due carousel tab"
```

---

## Task 8: Rating Color Coding

**Files:**
- Modify: `src/internal/library/adapters/dashboard.templ`
- Modify: `src/internal/library/adapters/album_detail.templ`

- [ ] **Step 1: Add ratingStateClass helper in dashboard.templ**

Add this helper function near the top of `dashboard.templ` (after the imports):

```go
// ratingStateClass returns the Tailwind class for the numeric rating value based on state and due status.
func ratingStateClass(album library.AlbumDTO) string {
	if album.RatingState == nil {
		return "text-base-content/20 hover:text-base-content/60"
	}
	switch album.RatingState.State {
	case "stalled":
		return "text-error/50 hover:text-error/70"
	case "finalized":
		return "text-primary hover:text-primary/80"
	case "provisional":
		if album.RatingState.IsRerateDue() {
			return "text-warning hover:text-warning/80"
		}
		return "text-base-content/40 hover:text-base-content/60"
	default:
		return "text-primary hover:text-primary/80"
	}
}
```

- [ ] **Step 2: Update AlbumListRating to use state-based colors**

Replace the existing `AlbumListRating` template. The key change is replacing the hardcoded `text-primary` class with `ratingStateClass(album)`, and adding a Stalled badge:

```go
templ AlbumListRating(album library.AlbumDTO, isOobSwap bool) {
	if album.Rating != nil && album.Rating.Rating != nil {
		<div
			id={ GetAlbumListRatingID(album.ID) }
			data-testid="album-row-rating"
			class={ "min-w-14 h-full flex flex-col items-center justify-center text-4xl font-bold tabular-nums cursor-pointer select-none transition-colors", ratingStateClass(album) }
			hx-get={ fmt.Sprintf("/app/review/rating-recommender?albumId=%s", album.ID) }
			hx-trigger="click"
			hx-swap="none"
			if isOobSwap {
				hx-swap-oob="true"
			}
		>
			{ fmt.Sprintf("%.1f", *album.Rating.Rating) }
			if album.RatingState != nil && album.RatingState.State == "stalled" {
				<span class="text-xs font-normal badge badge-xs badge-error badge-soft">Stalled</span>
			}
		</div>
	} else {
		<div
			id={ GetAlbumListRatingID(album.ID) }
			data-testid="album-row-rating"
			class="min-w-14 h-full flex items-center justify-center text-4xl font-bold tabular-nums cursor-pointer select-none text-base-content/20 hover:text-base-content/60 transition-colors"
			hx-get={ fmt.Sprintf("/app/review/rating-recommender?albumId=%s", album.ID) }
			hx-trigger="click"
			hx-swap="none"
			if isOobSwap {
				hx-swap-oob="true"
			}
		>
			--
		</div>
	}
}
```

- [ ] **Step 3: Update AlbumRating (badge on detail page) to use state colors**

In `dashboard.templ`, update `AlbumRating` to pick badge variant based on state:

```go
func ratingBadgeClass(album library.AlbumDTO) string {
	if album.RatingState == nil {
		return "badge badge-soft badge-primary text-nowrap cursor-pointer"
	}
	switch album.RatingState.State {
	case "stalled":
		return "badge badge-soft badge-error text-nowrap cursor-pointer opacity-60"
	case "finalized":
		return "badge badge-soft badge-primary text-nowrap cursor-pointer"
	case "provisional":
		if album.RatingState.IsRerateDue() {
			return "badge badge-soft badge-warning text-nowrap cursor-pointer"
		}
		return "badge badge-soft badge-ghost text-nowrap cursor-pointer opacity-50"
	default:
		return "badge badge-soft badge-primary text-nowrap cursor-pointer"
	}
}

templ AlbumRating(album library.AlbumDTO, isOobSwap bool) {
	if album.Rating != nil && album.Rating.Rating != nil {
		<div
			id={ GetAlbumRatingID(album.ID) }
			data-testid="album-row-rating"
			class={ ratingBadgeClass(album) }
			hx-get={ fmt.Sprintf("/app/review/rating-recommender?albumId=%s", album.ID) }
			hx-trigger="click"
			hx-swap="none"
			if isOobSwap {
				hx-swap-oob="true"
			}
		>
			if album.RatingState != nil && album.RatingState.State == "stalled" {
				Stalled —
			}
			{ fmt.Sprintf("%.1f", *album.Rating.Rating) } - { string(review.GetRatingLabel(*album.Rating.Rating)) }
		</div>
	} else {
		<button
			id={ GetAlbumRatingID(album.ID) }
			data-testid="album-row-rating"
			class="btn btn-xs btn-ghost opacity-20 hover:opacity-100 cursor-pointer text-nowrap"
			hx-get={ fmt.Sprintf("/app/review/rating-recommender?albumId=%s", album.ID) }
			hx-trigger="click"
			hx-swap="none"
			if isOobSwap {
				hx-swap-oob="true"
			}
		>
			Rate
		</button>
	}
}
```

- [ ] **Step 4: Update AlbumRatingHistory to show state badge per entry**

In `album_detail.templ`, inside the rating history entry loop, add a state badge after the score:

```go
<div class="flex items-center gap-2">
    <span class="badge badge-soft badge-primary text-nowrap" data-testid="rating-history-score">
        { fmt.Sprintf("%.4g", *entry.Rating) } - { string(review.GetRatingLabel(*entry.Rating)) }
    </span>
    if entry.State != nil {
        <span class="badge badge-xs badge-ghost text-nowrap">{ string(*entry.State) }</span>
    }
    <span class="text-xs text-base-content/40">{ entry.CreatedAt.Format("Jan 2, 2006") }</span>
</div>
```

- [ ] **Step 5: Build templates and verify**

```bash
cd /Users/shmoopy/workshop/projects/wax && task build/templ && go build ./src/... 2>&1
```

- [ ] **Step 6: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/library/adapters/ && git commit -m "feat(ui): state-based rating color coding and stalled badge"
```

---

## Task 9: New Questionnaire + Rerate Prompt UI

**Files:**
- Rewrite: `src/internal/review/adapters/rating.templ`

- [ ] **Step 1: Write the new rating.templ**

Replace the entire file:

```go
package adapters

import (
	"fmt"
	"github.com/alecdray/wax/src/internal/core/templates"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/review"
	"strconv"
)

const RatingModalId = "rating-modal"

// --- Shared confirm form helpers (unchanged from before) ---

func ratingConfirmAlpineData(rating *float64) string {
	initial := "''"
	if rating != nil {
		initial = strconv.FormatFloat(*rating, 'f', -1, 64)
	}
	key := "["
	for i, entry := range review.RatingKey {
		if i > 0 {
			key += ","
		}
		key += fmt.Sprintf("[%s,%s,'%s']",
			strconv.FormatFloat(entry.MinValue, 'f', -1, 64),
			strconv.FormatFloat(entry.MaxValue, 'f', -1, 64),
			string(entry.Label),
		)
	}
	key += "]"
	return fmt.Sprintf(`{rating: %s, getRatingLabel(v){const n=parseFloat(v);if(isNaN(n))return'';const c=Math.min(10,Math.max(0,n));const k=%s;for(const[mn,mx,l]of k)if(c>=mn&&c<=mx)return l;return k[k.length-1][2]}}`, initial, key)
}

// --- Base questions form ---

templ BaseQuestionRadio(q review.BaseQuestion, opt review.QuestionOption) {
	<label class="flex items-center gap-3 cursor-pointer">
		<input
			type="radio"
			name={ string(q.Key) }
			value={ strconv.Itoa(opt.Value) }
			class="radio radio-primary radio-sm"
			required
		/>
		<span class="text-base-content/70">{ opt.Label }</span>
	</label>
}

templ BaseQuestionFieldset(q review.BaseQuestion) {
	<fieldset class="flex flex-col gap-3">
		<legend class="text-base-content font-medium mb-2">{ q.Question }</legend>
		for _, opt := range q.Options {
			@BaseQuestionRadio(q, opt)
		}
	</fieldset>
}

// BaseQuestionsForm renders the base questions step.
// mode is "provisional" or "finalized". albumId identifies the album.
templ BaseQuestionsForm(albumId string, mode review.RatingMode, questions review.BaseQuestions) {
	<form
		data-testid="rating-questionnaire"
		hx-post={ fmt.Sprintf("/app/review/rating-recommender/questions?albumId=%s&mode=%s", albumId, string(mode)) }
		hx-swap="outerHTML"
		class="flex flex-col gap-8"
	>
		for _, q := range questions {
			if mode == review.RatingModeProvisional && q.Key == review.QuestionReturnRate {
				// skip return rate in provisional mode
			} else {
				@BaseQuestionFieldset(q)
			}
		}
		<button class="btn btn-primary" type="submit" data-testid="rating-calculate">Next</button>
	</form>
}

// --- Modifiers form ---

templ ModifierRadio(m review.Modifier, opt review.ModifierOption) {
	<label class="flex items-center gap-3 cursor-pointer">
		<input
			type="radio"
			name={ string(m.Key) }
			value={ strconv.Itoa(opt.Value) }
			class="radio radio-primary radio-sm"
			required
		/>
		<span class="text-base-content/70">{ opt.Label }</span>
	</label>
}

templ ModifierFieldset(m review.Modifier) {
	<fieldset class="flex flex-col gap-3">
		<legend class="text-base-content font-medium mb-2">{ m.Label }</legend>
		for _, opt := range m.Options {
			@ModifierRadio(m, opt)
		}
	</fieldset>
}

// ModifiersForm renders the modifiers step.
// Hidden inputs carry the mode and base score forward.
templ ModifiersForm(albumId string, mode review.RatingMode, baseScore float64, questionValues map[string]string) {
	<form
		data-testid="rating-modifiers"
		hx-post={ fmt.Sprintf("/app/review/rating-recommender/modifiers?albumId=%s", albumId) }
		hx-swap="outerHTML"
		class="flex flex-col gap-8"
	>
		<input type="hidden" name="mode" value={ string(mode) }/>
		<input type="hidden" name="base_score" value={ strconv.FormatFloat(baseScore, 'f', 2, 64) }/>
		for k, v := range questionValues {
			<input type="hidden" name={ k } value={ v }/>
		}
		<p class="text-xs text-base-content/50 uppercase tracking-widest font-semibold">Quick gut checks</p>
		for _, m := range review.AllModifiers {
			@ModifierFieldset(m)
		}
		<button class="btn btn-primary" type="submit" data-testid="rating-modifiers-submit">Calculate Score</button>
	</form>
}

// --- Confidence interstitial ---

templ ConfidenceInterstitial(albumId string, mode review.RatingMode, finalScore float64, formValues map[string]string) {
	<div data-testid="rating-confidence" class="flex flex-col gap-6">
		<div class="alert alert-warning">
			<p class="text-sm">Your answers contain some contradictions. Would you like to review them, or proceed with the calculated score?</p>
		</div>
		<div class="flex flex-col gap-3">
			<a
				class="btn btn-ghost"
				hx-get={ fmt.Sprintf("/app/review/rating-recommender/questions?albumId=%s&mode=%s", albumId, string(mode)) }
				hx-swap="outerHTML"
				hx-target="closest [data-testid='rating-confidence']"
				data-testid="rating-confidence-review"
			>Review answers</a>
			<form
				hx-post={ fmt.Sprintf("/app/review/rating-recommender/confirm?albumId=%s", albumId) }
				hx-swap="outerHTML"
				class="contents"
			>
				<input type="hidden" name="mode" value={ string(mode) }/>
				<input type="hidden" name="final_score" value={ strconv.FormatFloat(finalScore, 'f', 2, 64) }/>
				<button class="btn btn-primary" type="submit" data-testid="rating-confidence-proceed">Proceed with { fmt.Sprintf("%.1f", finalScore) }</button>
			</form>
		</div>
	</div>
}

// --- Confirm form ---

templ RatingConfirmForm(album library.AlbumDTO, mode review.RatingMode, rating *float64) {
	<form
		data-testid="rating-confirm"
		class="flex flex-col gap-4"
		hx-ext="morph"
		hx-post={ fmt.Sprintf("/app/review/rating-recommender/rating?albumId=%s", album.ID) }
	>
		<input type="hidden" name="mode" value={ string(mode) }/>
		<div class="flex items-center gap-2">
			<p class="max-w-[75%] text-ellipsis text-nowrap overflow-hidden">{ album.Title }</p>
			<div class="flex gap-1">
				<button
					class="btn btn-ghost btn-sm btn-square"
					type="button"
					hx-get={ fmt.Sprintf("/app/review/rating-recommender/questions?albumId=%s&mode=%s", album.ID, string(mode)) }
					hx-target="closest form"
					hx-swap="outerHTML"
				>
					@templates.QuestionMarkIcon(templates.IconProps{})
				</button>
			</div>
		</div>
		<fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-full border px-4">
			<legend class="fieldset-legend">Rating</legend>
			<label class="input w-full" x-data={ ratingConfirmAlpineData(rating) }>
				<input
					name="rating"
					data-testid="rating-input"
					type="number"
					min="0"
					max="10"
					class="grow"
					step="0.1"
					placeholder="Enter your rating"
					if rating != nil {
						value={ fmt.Sprintf("%.1f", *rating) }
					}
					required
					x-model="rating"
					autofocus
				/>
				<span class="badge badge-primary badge-soft badge-s" x-show="rating !== ''" x-text="getRatingLabel(rating)"></span>
			</label>
			<div class="collapse collapse-arrow">
				<input type="checkbox"/>
				<div class="collapse-title text-xs">Rating key</div>
				<div class="collapse-content">
					<table class="table table-xs">
						<tbody>
							for _, entry := range review.RatingKey {
								<tr>
									<td class="tabular-nums w-px whitespace-nowrap">{ strconv.FormatFloat(entry.MinValue, 'f', -1, 64) }</td>
									<td>{ string(entry.Label) }</td>
								</tr>
							}
						</tbody>
					</table>
				</div>
			</div>
			@RatingConfirmError("")
		</fieldset>
		<fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-full border px-4">
			<legend class="fieldset-legend">Note <span class="font-normal text-base-content/40">(optional)</span></legend>
			<textarea
				name="note"
				data-testid="rating-note"
				class="textarea w-full min-h-20"
				placeholder="Add a note about this rating..."
				maxlength="2000"
			></textarea>
		</fieldset>
		<button
			class="btn btn-primary w-full"
			type="submit"
			data-testid="rating-lock-in"
			hx-target-error="next .error"
		>Lock in</button>
		<p class="label error text-error"></p>
	</form>
}

templ RatingConfirmError(text string) {
	<p class="label error text-error">{ text }</p>
}

// --- Rerate prompt ---

type ReratePromptProps struct {
	Album   library.AlbumDTO
	IsDue   bool
	IsStalled bool
}

templ ReratePrompt(props ReratePromptProps) {
	<div data-testid="rerate-prompt" class="flex flex-col gap-4">
		<p class="text-sm text-base-content/70 truncate">{ props.Album.Title }</p>
		if props.Album.Rating != nil && props.Album.Rating.Rating != nil {
			<p class="text-xs text-base-content/40">
				Current score: <span class="font-bold tabular-nums">{ fmt.Sprintf("%.1f", *props.Album.Rating.Rating) }</span>
				if props.Album.RatingState != nil {
					<span class="badge badge-xs badge-ghost ml-1">{ string(props.Album.RatingState.State) }</span>
				}
			</p>
		}
		<div class="flex flex-col gap-2">
			<a
				class="btn btn-primary"
				hx-get={ fmt.Sprintf("/app/review/rating-recommender/questions?albumId=%s&mode=finalized", props.Album.ID) }
				hx-target="closest [data-testid='rerate-prompt']"
				hx-swap="outerHTML"
				data-testid="rerate-rate-now"
			>Rate now</a>
			if !props.IsStalled {
				if props.IsDue {
					<form hx-post={ fmt.Sprintf("/app/review/rating/snooze?albumId=%s", props.Album.ID) } hx-swap="none" class="contents">
						<button class="btn btn-ghost" type="submit" data-testid="rerate-snooze">Snooze 1 week</button>
					</form>
				} else {
					<a
						class="btn btn-ghost"
						hx-get={ fmt.Sprintf("/app/review/rating-recommender/questions?albumId=%s&mode=provisional", props.Album.ID) }
						hx-target="closest [data-testid='rerate-prompt']"
						hx-swap="outerHTML"
						data-testid="rerate-update-provisional"
					>Update provisional</a>
				}
			}
		</div>
	</div>
}

// --- Modal wrapper ---

type RatingModalContent int

const (
	RatingModalContentQuestions    RatingModalContent = iota
	RatingModalContentReratePrompt RatingModalContent = iota
	RatingModalContentConfirm      RatingModalContent = iota
)

type RatingModalProps struct {
	Album         library.AlbumDTO
	Mode          review.RatingMode
	Rating        *float64
	ContentType   RatingModalContent
	RerateIsDue   bool
	RerateIsStalled bool
}

templ RatingModal(props RatingModalProps) {
	switch props.ContentType {
	case RatingModalContentQuestions:
		@templates.Modal(RatingModalId, templates.ModalProps{
			ModalContent: BaseQuestionsForm(props.Album.ID, props.Mode, review.AllBaseQuestions),
		})
	case RatingModalContentReratePrompt:
		@templates.Modal(RatingModalId, templates.ModalProps{
			ModalContent: ReratePrompt(ReratePromptProps{
				Album:     props.Album,
				IsDue:     props.RerateIsDue,
				IsStalled: props.RerateIsStalled,
			}),
		})
	default:
		@templates.Modal(RatingModalId, templates.ModalProps{
			ModalContent: RatingConfirmForm(props.Album, props.Mode, props.Rating),
		})
	}
}

templ CloseRatingModal() {
	@templates.ForceCloseModal(RatingModalId)
}
```

- [ ] **Step 2: Build templates**

```bash
cd /Users/shmoopy/workshop/projects/wax && task build/templ 2>&1
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/shmoopy/workshop/projects/wax && go build ./src/... 2>&1
```

Expected: errors in `review/adapters/http.go` — the old template call sites no longer exist. These are fixed in the next task.

- [ ] **Step 4: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/review/adapters/rating.templ src/internal/review/adapters/rating_templ.go && git commit -m "feat(review): new multi-step questionnaire and rerate prompt templates"
```

---

## Task 10: Update HTTP Handlers

**Files:**
- Rewrite: `src/internal/review/adapters/http.go`
- Modify: `src/internal/server/server.go`

- [ ] **Step 1: Rewrite http.go**

Replace the entire file:

```go
package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
	libAdapters "github.com/alecdray/wax/src/internal/library/adapters"
	"github.com/alecdray/wax/src/internal/review"
)

type HttpHandler struct {
	libraryService *library.Service
	reviewService  *review.Service
}

func NewHttpHandler(libraryService *library.Service, reviewService *review.Service) *HttpHandler {
	return &HttpHandler{
		libraryService: libraryService,
		reviewService:  reviewService,
	}
}

// getAlbum is a helper to look up an album by query param, returning early on error.
func (h *HttpHandler) getAlbum(ctx contextx.ContextX, w http.ResponseWriter, userID, albumID string) (*library.AlbumDTO, bool) {
	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return nil, false
	}
	return album, true
}

// GetRatingRecommender opens the rating modal. Behaviour depends on album state:
//   - No state (never rated): opens base questions form in provisional mode
//   - Provisional or stalled: opens rerate prompt
//   - Finalized: opens confirm screen (pre-filled)
func (h *HttpHandler) GetRatingRecommender(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	props := RatingModalProps{Album: *album}

	if album.RatingState == nil {
		// Never rated — start the provisional questionnaire.
		props.ContentType = RatingModalContentQuestions
		props.Mode = review.RatingModeProvisional
	} else if album.RatingState.State == review.RatingStateFinalized {
		// Finalized — open confirm screen for manual re-rating.
		props.ContentType = RatingModalContentConfirm
		props.Mode = review.RatingModeFinalized
		if album.Rating != nil {
			props.Rating = album.Rating.Rating
		}
	} else {
		// Provisional or stalled — show rerate prompt.
		props.ContentType = RatingModalContentReratePrompt
		props.RerateIsDue = album.RatingState.IsRerateDue()
		props.RerateIsStalled = album.RatingState.State == review.RatingStateStalled
	}

	if err := RatingModal(props).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetRatingRecommenderQuestions renders the base questions form.
func (h *HttpHandler) GetRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	mode := review.RatingMode(r.URL.Query().Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	if err := BaseQuestionsForm(albumID, mode, review.AllBaseQuestions).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// SubmitRatingRecommenderQuestions processes base question answers and returns the modifier form.
func (h *HttpHandler) SubmitRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")
	mode := review.RatingMode(r.URL.Query().Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	questions := make(review.BaseQuestions, len(review.AllBaseQuestions))
	copy(questions, review.AllBaseQuestions)

	questionValues := make(map[string]string)
	for i, q := range questions {
		if mode == review.RatingModeProvisional && q.Key == review.QuestionReturnRate {
			continue
		}
		rawVal := r.Form.Get(string(q.Key))
		if rawVal == "" {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("missing value for question %s", q.Key),
			})
			return
		}
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for %s: %w", q.Key, err)})
			return
		}
		questions[i] = q.WithValue(val)
		questionValues[string(q.Key)] = rawVal
	}

	baseScore := questions.Score(mode)

	if err := ModifiersForm(albumID, mode, baseScore, questionValues).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// SubmitModifiers processes modifier answers, checks for contradictions, returns confirm or confidence interstitial.
func (h *HttpHandler) SubmitModifiers(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	baseScore, err := strconv.ParseFloat(r.Form.Get("base_score"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid base_score: %w", err)})
		return
	}

	mods := make(review.Modifiers, len(review.AllModifiers))
	copy(mods, review.AllModifiers)
	for i, m := range mods {
		rawVal := r.Form.Get(string(m.Key))
		if rawVal == "" {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("missing value for modifier %s", m.Key)})
			return
		}
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for modifier %s: %w", m.Key, err)})
			return
		}
		mods[i] = m.WithValue(val)
	}

	finalScore := review.FinalScore(baseScore, mods.Adjustment())

	// Rebuild question values from hidden inputs for contradiction check.
	questions := make(review.BaseQuestions, len(review.AllBaseQuestions))
	copy(questions, review.AllBaseQuestions)
	for i, q := range questions {
		rawVal := r.Form.Get(string(q.Key))
		if rawVal != "" {
			val, _ := strconv.Atoi(rawVal)
			questions[i] = q.WithValue(val)
		}
	}

	if review.DetectContradictions(questions, mods, baseScore, mode) {
		// Collect all form values to pass through the confidence interstitial.
		allValues := make(map[string]string)
		for _, q := range questions {
			k := string(q.Key)
			if v := r.Form.Get(k); v != "" {
				allValues[k] = v
			}
		}
		for _, m := range mods {
			k := string(m.Key)
			allValues[k] = r.Form.Get(k)
		}
		if err := ConfidenceInterstitial(albumID, mode, finalScore, allValues).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		}
		return
	}

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	if err := RatingConfirmForm(*album, mode, &finalScore).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetRatingConfirm renders the confirm form with a pre-computed score (from the confidence interstitial "Proceed" button).
func (h *HttpHandler) GetRatingConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	finalScore, err := strconv.ParseFloat(r.Form.Get("final_score"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid final_score: %w", err)})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	if err := RatingConfirmForm(*album, mode, &finalScore).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// SubmitRatingRecommenderRating saves the rating and upserts the lifecycle state.
func (h *HttpHandler) SubmitRatingRecommenderRating(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	ratingVal, err := strconv.ParseFloat(r.Form.Get("rating"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid rating: %w", err)})
		return
	}

	note := r.Form.Get("note")
	if len(note) > 2000 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("note exceeds 2000 character limit")})
		return
	}

	// Map mode to state for the log entry.
	logState := review.RatingStateProvisional
	if mode == review.RatingModeFinalized {
		logState = review.RatingStateFinalized
	}

	albumRating, err := h.reviewService.AddRating(ctx, userID, albumID, ratingVal, note, logState)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to add rating: %w", err)})
		return
	}
	_ = albumRating

	// Upsert lifecycle state.
	currentState, err := h.reviewService.GetRatingState(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}

	if mode == review.RatingModeFinalized {
		if currentState == nil {
			// Edge case: finalize with no prior state — create and immediately finalize.
			if _, err := h.reviewService.CreateRatingState(ctx, userID, albumID); err != nil {
				httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
				return
			}
		}
		if _, err := h.reviewService.FinalizeRating(ctx, userID, albumID, currentState); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
			return
		}
	} else if mode == review.RatingModeProvisional && currentState == nil {
		// First ever rating — create provisional state.
		if _, err := h.reviewService.CreateRatingState(ctx, userID, albumID); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
			return
		}
	}
	// mode=provisional + existing state → update provisional, state row unchanged.

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get album: %w", err)})
		return
	}

	if err := CloseRatingModal().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRatingHistory(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
}

// SnoozeRating handles a snooze request: increments snooze count and bumps next_rerate_at.
func (h *HttpHandler) SnoozeRating(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if _, err := h.reviewService.SnoozeRating(ctx, userID, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to snooze: %w", err)})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	if err := CloseRatingModal().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
}

// DeleteRatingLogEntry removes a single rating history entry.
func (h *HttpHandler) DeleteRatingLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get user ID: %w", err)})
		return
	}

	entryID := r.PathValue("id")
	if entryID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing entry ID")})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := h.reviewService.DeleteRatingEntry(ctx, userID, entryID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to delete: %w", err)})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get album: %w", err)})
		return
	}

	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRatingHistory(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
}
```

- [ ] **Step 2: Register new routes in server.go**

In `src/internal/server/server.go`, find the review route block (around line 165) and add the new routes:

```go
appMux.Handle("GET /app/review/rating-recommender", httpx.HandlerFunc(reviewHandler.GetRatingRecommender))
appMux.Handle("GET /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.GetRatingRecommenderQuestions))
appMux.Handle("POST /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderQuestions))
appMux.Handle("POST /app/review/rating-recommender/modifiers", httpx.HandlerFunc(reviewHandler.SubmitModifiers))
appMux.Handle("POST /app/review/rating-recommender/confirm", httpx.HandlerFunc(reviewHandler.GetRatingConfirm))
appMux.Handle("POST /app/review/rating-recommender/rating", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderRating))
appMux.Handle("POST /app/review/rating/snooze", httpx.HandlerFunc(reviewHandler.SnoozeRating))
appMux.Handle("DELETE /app/review/rating-log/{id}", httpx.HandlerFunc(reviewHandler.DeleteRatingLogEntry))
```

- [ ] **Step 3: Wire reviewService into library.Service if needed**

Check if `library.Service` already holds a `reviewService` (you may have done this in Task 6). If not, look at the `Service` struct in `library/service.go`, add the field, and update `NewService` + the call site in `server.go`. The `server.go` already creates `review.NewService(db)` — pass that instance to `library.NewService`.

- [ ] **Step 4: Full build and test**

```bash
cd /Users/shmoopy/workshop/projects/wax && task build/templ && go build ./src/... 2>&1
```

```bash
cd /Users/shmoopy/workshop/projects/wax && go test ./src/... -v 2>&1
```

Expected: all tests PASS. Fix any compilation or test failures before committing.

- [ ] **Step 5: Commit**

```bash
cd /Users/shmoopy/workshop/projects/wax && git add src/internal/review/adapters/ src/internal/server/server.go && git commit -m "feat(review): new HTTP handlers for multi-step questionnaire, modifiers, rerate, and snooze"
```

---

## Self-Review Checklist

Before declaring done, verify against spec:

- [ ] Scoring engine: 6 questions, 3 modifiers, formula matches spec — ✅ Task 1
- [ ] Equal default weights, configurable as named constants — ✅ Task 1
- [ ] Provisional: Return Rate excluded, cap at 8.0 — ✅ Task 1
- [ ] Confidence flag: 3 contradiction patterns, non-blocking — ✅ Tasks 1 + 9
- [ ] DB: `album_rating_state` table + `state` column on log — ✅ Tasks 3 + 4
- [ ] State transitions: provisional → finalized, snooze → stalled after 3 — ✅ Task 5
- [ ] Snooze: 1 week per snooze, max 3 before stalled — ✅ Task 5
- [ ] Update provisional: no change to next_rerate_at — ✅ Task 10 handler
- [ ] Finalized albums: re-ratable (finalized → finalized) — ✅ Task 10 (mode=finalized path)
- [ ] Rerate prompt: 2 options, context-aware (due: snooze / not due: update provisional) — ✅ Task 9
- [ ] Stalled: Rate now only — ✅ Task 9
- [ ] Color coding: finalized/provisional/stalled/rerate-due — ✅ Task 8
- [ ] Stalled badge in list — ✅ Task 8
- [ ] Rerate Due carousel tab (provisional+due + stalled) — ✅ Task 7
- [ ] State badge in rating history — ✅ Task 8
