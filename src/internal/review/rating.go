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

// Score computes the score (0–10) for the answered questions.
// Unanswered questions (Value == 0) are excluded from both sums.
func (qs BaseQuestions) Score() float64 {
	var weightedSum, totalWeight float64
	for _, q := range qs {
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

// FinalScore clamps and rounds the score to one decimal place.
func FinalScore(baseScore float64) float64 {
	return math.Round(utils.Clamp(baseScore, 0.0, 10.0)*10) / 10
}

// DetectContradictions returns true if the answers contain internally contradictory signals.
// Only checked in finalized mode: high Sonic Pleasure but low Return Rate.
func DetectContradictions(qs BaseQuestions, mode RatingMode) bool {
	if mode != RatingModeFinalized {
		return false
	}
	qByKey := make(map[BaseQuestionKey]int, len(qs))
	for _, q := range qs {
		qByKey[q.Key] = q.Value
	}
	sp := qByKey[QuestionSonicPleasure]
	rr := qByKey[QuestionReturnRate]
	return sp >= 4 && rr <= 2
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
