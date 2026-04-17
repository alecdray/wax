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
// Hard questions are near pass/fail requirements (weight 2).
// Soft questions measure degree of quality (weight 1).
const (
	RatingWeightHard = 2.0
	RatingWeightSoft = 1.0

	// ModifierMaxSwing is the maximum total modifier adjustment (positive or negative).
	ModifierMaxSwing = 0.75

	// ProvisionalScoreCap is the maximum score a provisional rating can produce.
	ProvisionalScoreCap = 8.0
)

// BaseQuestionKey identifies a base question.
type BaseQuestionKey string

const (
	QuestionReturnRate         BaseQuestionKey = "return_rate"
	QuestionTrackQuality       BaseQuestionKey = "track_quality"
	QuestionCohesion           BaseQuestionKey = "cohesion"
	QuestionEmotionalResonance BaseQuestionKey = "emotional_resonance"
	QuestionSonicPleasure      BaseQuestionKey = "sonic_pleasure"
	QuestionShelfTest          BaseQuestionKey = "shelf_test"
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
// In provisional mode, ReturnRate and ShelfTest are excluded (attachment questions requiring lived experience).
// The provisional cap is applied in FinalScore.
func (qs BaseQuestions) Score(mode RatingMode) float64 {
	var weightedSum, totalWeight float64
	for _, q := range qs {
		if mode == RatingModeProvisional && (q.Key == QuestionReturnRate || q.Key == QuestionShelfTest) {
			continue
		}
		if q.Value == 0 {
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
	return math.Round(base*10) / 10
}

// AllBaseQuestions is the canonical ordered list of base questions.
var likertOptions = []QuestionOption{
	{1, "Strongly disagree"},
	{2, "Disagree"},
	{3, "Neutral"},
	{4, "Agree"},
	{5, "Strongly agree"},
}

var AllBaseQuestions = BaseQuestions{
	{
		Key:      QuestionReturnRate,
		Question: "I will keep coming back to this record",
		Options:  likertOptions,
		Weight:   RatingWeightHard,
	},
	{
		Key:      QuestionTrackQuality,
		Question: "The tracks on this record consistently land",
		Options:  likertOptions,
		Weight:   RatingWeightSoft,
	},
	{
		Key:      QuestionCohesion,
		Question: "This record works as a complete piece",
		Options:  likertOptions,
		Weight:   RatingWeightSoft,
	},
	{
		Key:      QuestionEmotionalResonance,
		Question: "This record makes me feel something",
		Options:  likertOptions,
		Weight:   RatingWeightSoft,
	},
	{
		Key:      QuestionSonicPleasure,
		Question: "I enjoy listening to this record",
		Options:  likertOptions,
		Weight:   RatingWeightHard,
	},
	{
		Key:      QuestionShelfTest,
		Question: "I would care if I had to permanently delete this record",
		Options:  likertOptions,
		Weight:   RatingWeightHard,
	},
}

// ModifierKey identifies a modifier.
type ModifierKey string

const (
	ModifierLifeAssociation ModifierKey = "life_association"
	ModifierInterest        ModifierKey = "interest"
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
		Key:   ModifierLifeAssociation,
		Label: "Life Association",
		Options: []ModifierOption{
			{1, "Meaningful association"},
			{0, "No strong history"},
			{-1, "Avoidant association"},
		},
	},
	{
		Key:   ModifierInterest,
		Label: "Interest",
		Options: []ModifierOption{
			{1, "Yes"},
			{0, "Neutral"},
			{-1, "Not particularly"},
		},
	},
}

// FinalScore clamps and rounds the combined base score + modifier adjustment.
// In provisional mode, the result is additionally capped at ProvisionalScoreCap.
func FinalScore(baseScore, modifierAdjustment float64, mode RatingMode) float64 {
	combined := baseScore + modifierAdjustment
	if mode == RatingModeProvisional {
		combined = math.Min(combined, ProvisionalScoreCap)
	}
	return math.Round(utils.Clamp(combined, 0.0, 10.0)*10) / 10
}

// DetectContradictions returns true if the answers contain internally contradictory signals.
// Contradiction 1 (finalized only): high Emotional Resonance AND Sonic Pleasure, but low Return Rate.
// Contradiction 2: high base score but all modifiers negative.
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
	if baseScore >= 7.0 && len(mods) > 0 {
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
	{4.0, 4.9, RatingLabelNotForMe},
	{5.0, 5.9, RatingLabelLukewarm},
	{6.0, 6.9, RatingLabelSolid},
	{7.0, 7.9, RatingLabelRecommended},
	{8.0, 8.9, RatingLabelEssential},
	{9.0, 9.9, RatingLabelInstantClassic},
	{10.0, 10.0, RatingLabelMasterpiece},
}

func GetRatingLabel(rating float64) RatingLabel {
	clamped := utils.Clamp(rating, 0, 10)
	for i := len(RatingKey) - 1; i >= 0; i-- {
		if clamped >= RatingKey[i].MinValue {
			return RatingKey[i].Label
		}
	}
	return RatingKey[0].Label
}
