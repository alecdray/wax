package review

import (
	"math"
	"shmoopicks/src/internal/core/utils"
)

const (
	// Controls how much extreme answers (1 or 5) are amplified.
	// Lower values = more powerful extremes, higher values = closer to linear.
	// Valid range 0.0–1.0, where 1.0 is fully linear.
	scoreCurveExponent = 0.8

	// Minimum possible score — awarded when all answers are 1.
	scoreFloor = 2.0

	// Maximum possible score — awarded when all answers are 5.
	scoreCeiling = 10.0

	// Weighting for Q1: track-by-track consistency.
	// Controls how much skips and weak tracks pull the score down.
	ScoreWeightConsistency = 0.3

	// Weighting for Q2: emotional impact while listening.
	// Highest weight — the primary differentiator between good and great.
	ScoreWeightImpact = 0.4

	// Weighting for Q3: immediate gut reaction when the album ended.
	// Captures the overall impression beyond individual tracks.
	ScoreWeightGutCheck = 0.3
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
	curved := math.Copysign(math.Pow(math.Abs(normalized), scoreCurveExponent), normalized)
	return curved*2.0 + 3.0
}

func (qs RatingQuestion) WithValue(value int) RatingQuestion {
	qs.Value = value
	return qs
}

type RatingQuestions []RatingQuestion

func (qs RatingQuestions) Score() float64 {
	var raw float64
	for _, score := range qs {
		raw += score.CurvedValue() * score.Weight
	}
	score := scoreFloor + ((raw-1.0)/4.0)*(scoreCeiling-scoreFloor)
	return math.Round(score*10) / 10
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
		Weight: ScoreWeightConsistency,
	},
	{
		Key:      RatingQuestionImpact,
		Question: "How did this album make you feel while listening?",
		Options: []RatingQuestionOption{
			{1, "Bored or uncomfortable"},
			{2, "Mostly indifferent"},
			{3, "Engaged but nothing special"},
			{4, "Moved — emotionally or physically"},
			{5, "Completely absorbed, transported somewhere else"},
		},
		Weight: ScoreWeightImpact,
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
		Weight: ScoreWeightGutCheck,
	},
}

type ScoreLabel string

const (
	ScoreLabelDOA                ScoreLabel = "DOA"
	ScoreLabelNope               ScoreLabel = "Nope"
	ScoreLabelNotForMe           ScoreLabel = "Not For Me"
	ScoreLabelHasItsMoments      ScoreLabel = "Has Its Moments"
	ScoreLabelGoodNotGreat       ScoreLabel = "Good Not Great"
	ScoreLabelWouldRecommend     ScoreLabel = "Would Recommend"
	ScoreLabelEssentialListening ScoreLabel = "Essential Listening"
	ScoreLabelInstantClassic     ScoreLabel = "Instant Classic"
	ScoreLabelMasterpiece        ScoreLabel = "Masterpiece"
)

func GetScoreLabel(score float64) ScoreLabel {
	clappedScore := utils.Clamp(score, 0, 10)

	switch {
	case clappedScore < 3.0:
		return ScoreLabelDOA
	case clappedScore < 4.0:
		return ScoreLabelNope
	case clappedScore < 6.0:
		return ScoreLabelNotForMe
	case clappedScore < 6.6:
		return ScoreLabelHasItsMoments
	case clappedScore < 7.0:
		return ScoreLabelGoodNotGreat
	case clappedScore < 8.0:
		return ScoreLabelWouldRecommend
	case clappedScore < 9.0:
		return ScoreLabelEssentialListening
	case clappedScore < 10.0:
		return ScoreLabelInstantClassic
	default:
		return ScoreLabelMasterpiece
	}
}
