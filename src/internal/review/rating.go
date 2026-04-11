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
