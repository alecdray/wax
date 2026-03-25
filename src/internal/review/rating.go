package review

import (
	"math"

	"github.com/alecdray/wax/src/internal/core/utils"
)

const (
	// Controls how much extreme answers (1 or 5) are amplified.
	// Lower values = more powerful extremes, higher values = closer to linear.
	// Valid range 0.0–1.0, where 1.0 is fully linear.
	ratingCurveExponent = 1.55

	// Minimum possible rating — awarded when all answers are 1.
	ratingFloor = 1.5

	// Maximum possible rating — awarded when all answers are 5.
	ratingCeiling = 10.0

	// Weighting for Q1: track-by-track consistency.
	// Controls how much skips and weak tracks pull the rating down.
	RatingWeightConsistency = 0.3

	// Weighting for Q2: emotional impact while listening.
	// Highest weight — the primary differentiator between good and great.
	RatingWeightImpact = 0.4

	// Weighting for Q3: immediate gut reaction when the album ended.
	// Captures the overall impression beyond individual tracks.
	RatingWeightGutCheck = 0.3
)

type RatingQuestionKey string

const (
	RatingQuestionConsistency RatingQuestionKey = "consistency"
	RatingQuestionImpact      RatingQuestionKey = "impact"
	RatingQuestionGutCheck    RatingQuestionKey = "gut_check"
)

func (k RatingQuestionKey) String() string {
	return string(k)
}

type RatingQuestionOption struct {
	Value int
	Label string
}

type RatingQuestion struct {
	Key      RatingQuestionKey
	Question string
	Options  []RatingQuestionOption
	Value    int
	Weight   float64
}

func (qs RatingQuestion) CurvedValue() float64 {
	normalized := (float64(qs.Value) - 3.0) / 2.0
	curved := math.Copysign(math.Pow(math.Abs(normalized), ratingCurveExponent), normalized)
	return curved*2.0 + 3.0
}

func (qs RatingQuestion) WithValue(value int) RatingQuestion {
	qs.Value = value
	return qs
}

type RatingQuestions []RatingQuestion

func (qs RatingQuestions) Rating() float64 {
	var raw float64
	for _, question := range qs {
		raw += question.CurvedValue() * question.Weight
	}
	rating := ratingFloor + ((raw-1.0)/4.0)*(ratingCeiling-ratingFloor)
	return math.Round(rating*10) / 10
}

var RatingRecommenderQuestions RatingQuestions = RatingQuestions{
	{
		Key:      RatingQuestionConsistency,
		Question: "How would you describe the album track by track?",
		Options: []RatingQuestionOption{
			{1, "Almost every track is a skip"},
			{2, "More misses than hits"},
			{3, "Mixed — some great, some weak"},
			{4, "Mostly strong with a few weak spots"},
			{5, "No skips, front to back"},
		},
		Weight: RatingWeightConsistency,
	},
	{
		Key:      RatingQuestionImpact,
		Question: "How did this album make you feel while listening?",
		Options: []RatingQuestionOption{
			{1, "Bored or uncomfortable"},
			{2, "Mostly indifferent"},
			{3, "Engaged but nothing special"},
			{4, "Moved or deeply engaged"},
			{5, "Completely absorbed, transported somewhere else"},
		},
		Weight: RatingWeightImpact,
	},
	{
		Key:      RatingQuestionGutCheck,
		Question: "When the album ended, what was your immediate reaction?",
		Options: []RatingQuestionOption{
			{1, "Relief it was over"},
			{2, "Indifferent"},
			{3, "Satisfied"},
			{4, "Impressed"},
			{5, "I immediately restarted it"},
		},
		Weight: RatingWeightGutCheck,
	},
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

// RatingKeyEntry defines an inclusive rating range and its associated label.
type RatingKeyEntry struct {
	// MinValue is the lowest rating (inclusive) that maps to this label.
	MinValue float64
	// MaxValue is the highest rating (inclusive) that maps to this label.
	MaxValue float64
	// Label is the human-readable name for this rating range.
	Label RatingLabel
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
	clampedRating := utils.Clamp(rating, 0, 10)
	for _, entry := range RatingKey {
		if clampedRating >= entry.MinValue && clampedRating <= entry.MaxValue {
			return entry.Label
		}
	}
	return RatingKey[len(RatingKey)-1].Label
}
